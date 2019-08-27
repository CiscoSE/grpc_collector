package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

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
	var namebuf bytes.Buffer
	telemetry := &telemetry.Telemetry{}
	// Unmarshal binary data into struct
	err := proto.Unmarshal(data, telemetry)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	for _, gpbkv := range telemetry.DataGpbkv {
		// Define file map
		var fields map[string]interface{}

		// Produce metadata tags
		var tags map[string]string

		// Top-level field may have measurement timestamp, if not use message timestamp
		measured := gpbkv.Timestamp
		if measured == 0 {
			measured = telemetry.MsgTimestamp
		}

		timestamp := time.Unix(int64(measured/1000), int64(measured%1000)*1000000)

		// Populate tags and fields from toplevel GPBKV fields "keys" and "content"
		for _, field := range gpbkv.Fields {
			switch field.Name {
			case "keys":
				tags = make(map[string]string, len(field.Fields)+2)
				tags["Producer"] = telemetry.GetNodeIdStr()
				tags["Target"] = telemetry.GetSubscriptionIdStr()
				tags["EncodingPath"] = telemetry.EncodingPath
				tags["TimeStamp"] = timestamp.String()
				for _, subfield := range field.Fields {
					c.parseGPBKVField(subfield, &namebuf, telemetry.EncodingPath, timestamp, tags, nil)
				}
			case "content":
				fields = make(map[string]interface{}, len(field.Fields))
				for _, subfield := range field.Fields {
					c.parseGPBKVField(subfield, &namebuf, telemetry.EncodingPath, timestamp, tags, fields)
				}
			default:
				log.Printf("I! Unexpected top-level MDT field: %s", field.Name)
			}
		}

		// Print measurement
		if len(fields) > 0 && len(tags) > 0 && len(telemetry.EncodingPath) > 0 {

			log.Printf("\n**** New Telemetry message from %v ****", tags["Producer"])
			log.Printf("Tags: %v", tags)
			log.Printf("Fields: %v\n", fields)
			//log.Printf(telemetry.EncodingPath, fields, tags, timestamp)

		} else {
			fmt.Printf("I! Cisco MDT invalid field: encoding path or measurement empty")
		}
	}

}

// Recursively parse GPBKV field structure into fields or tags
func (c *DialOutServer) parseGPBKVField(field *telemetry.TelemetryField, namebuf *bytes.Buffer,
	path string, timestamp time.Time, tags map[string]string, fields map[string]interface{}) {

	namelen := namebuf.Len()
	if namelen > 0 {
		namebuf.WriteRune('/')
	}
	namebuf.WriteString(field.Name)

	// Decode Telemetry field value if set
	var value interface{}
	switch field.ValueByType.(type) {
	case *telemetry.TelemetryField_BytesValue:
		value = field.ValueByType.(*telemetry.TelemetryField_BytesValue).BytesValue
	case *telemetry.TelemetryField_StringValue:
		value = field.ValueByType.(*telemetry.TelemetryField_StringValue).StringValue
	case *telemetry.TelemetryField_BoolValue:
		value = field.ValueByType.(*telemetry.TelemetryField_BoolValue).BoolValue
	case *telemetry.TelemetryField_Uint32Value:
		value = field.ValueByType.(*telemetry.TelemetryField_Uint32Value).Uint32Value
	case *telemetry.TelemetryField_Uint64Value:
		value = field.ValueByType.(*telemetry.TelemetryField_Uint64Value).Uint64Value
	case *telemetry.TelemetryField_Sint32Value:
		value = field.ValueByType.(*telemetry.TelemetryField_Sint32Value).Sint32Value
	case *telemetry.TelemetryField_Sint64Value:
		value = field.ValueByType.(*telemetry.TelemetryField_Sint64Value).Sint64Value
	case *telemetry.TelemetryField_DoubleValue:
		value = field.ValueByType.(*telemetry.TelemetryField_DoubleValue).DoubleValue
	case *telemetry.TelemetryField_FloatValue:
		value = field.ValueByType.(*telemetry.TelemetryField_FloatValue).FloatValue
	}

	if value != nil {
		// Distinguish between tags (keys) and fields (data) to write to
		if fields != nil {
			fields[namebuf.String()] = value
		} else {
			tags[namebuf.String()] = fmt.Sprint(value)
		}
	}

	for _, subfield := range field.Fields {
		c.parseGPBKVField(subfield, namebuf, path, timestamp, tags, fields)
	}

	namebuf.Truncate(namelen)
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
