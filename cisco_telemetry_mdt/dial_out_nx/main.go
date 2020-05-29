package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/CiscoSE/grpc_collector/cisco_telemetry_mdt/dial_out_nx/nx_telemetry_proto/urib"
	dialout "github.com/CiscoSE/grpc_collector/cisco_telemetry_mdt/mdt_dialout"
	"github.com/CiscoSE/grpc_collector/cisco_telemetry_mdt/telemetry"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type DialOutServer struct {
	cancel context.CancelFunc
	ctx    context.Context
}

// MdtDialout RPC server method for grpc-dialout transport
func (c *DialOutServer) MdtDialout(stream dialout.GRPCMdtDialout_MdtDialoutServer) error {
	// Validate the context
	peer, peerOK := peer.FromContext(stream.Context())
	if peerOK {
		log.Printf("Accepted Cisco MDT GRPC dialout connection from %s", peer.Addr)
	}

	for {
		// Process every packet until EOF or until no more data
		packet, err := stream.Recv()
		if err != nil {
			if err != io.EOF && c.ctx.Err() == nil {
				fmt.Printf("E! GRPC dialout receive error: %v", err)
			}
			break
		}

		if len(packet.Data) == 0 && len(packet.Errors) != 0 {
			log.Printf("No more data")
			break
		}
		// Process data
		c.handleTelemetry(packet.Data)
	}

	if peerOK {
		log.Printf("Closed Cisco MDT GRPC dialout connection from %s", peer.Addr)
	}

	return nil
}

// Handle telemetry packet from any transport, decode and add as measurement
func (c *DialOutServer) handleTelemetry(data []byte) {
	//
	//var namebuf bytes.Buffer
	message := &telemetry.Telemetry{}
	// Unmarshal binary data into struct
	err := proto.Unmarshal(data, message)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}
	log.Printf("***** New message from %v ***** \n", message.GetNodeId())

	for _, row := range message.GetDataGpb().GetRow() {

		// Top-level field may have measurement timestamp
		measured := row.GetTimestamp()

		timestamp := time.Unix(int64(measured/1000), int64(measured%1000)*1000000)
		content := row.GetContent()
		routeL3 := new(urib.NxL3RouteProto)

		err = proto.Unmarshal(content, routeL3)
		if err != nil {
			log.Fatalf("Could decode Content: %v\n", err)
		}

		log.Printf("Row timestamp: %d", timestamp)
		log.Printf("Row content: %v", routeL3)
		log.Printf("Row address: %s", routeL3.GetAddress())
	}

}
func main() {
	// Create new server struct
	c := &DialOutServer{}

	// Add context
	c.ctx, c.cancel = context.WithCancel(context.Background())

	// Configure protocol and ports
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	log.Printf("Listening in port: %d", *port)
	if err != nil {
		log.Fatalf("Failed to listen in configured port: %v", err)
	}

	// Create empty option
	var opts []grpc.ServerOption

	// Create new gRPC server
	grpcServer := grpc.NewServer(opts...)

	// Register server to gRPC calls
	dialout.RegisterGRPCMdtDialoutServer(grpcServer, c)

	// Start server
	grpcServer.Serve(lis)

}
