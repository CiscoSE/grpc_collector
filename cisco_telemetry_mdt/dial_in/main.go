/*
Dial-in to specified router in specified subscription
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/golang/protobuf/proto"
	xr "github.com/nleiva/xrgrpc"

	lldp "github.com/CiscoSE/grpc_collector/cisco_telemetry_mdt/telemetry/cisco_ios_xr_ethernet_lldp_oper/lldp/nodes/node/neighbors/summaries/summary"
	"github.com/nleiva/xrgrpc/proto/telemetry"
)

func main() {

	// Determine the ID for first the transaction.
	var id int64 = 1000

	// Manually specify target parameters.
	router1, err := xr.BuildRouter(
		xr.WithUsername("YOUR_USER"),         // e.g. admin
		xr.WithPassword("YOUR_PASS"),         // e.g. cisco123
		xr.WithHost("YOUR_DEVICE:GRPC_PORT"), // e.g. 192.168.0.1:57344
		xr.WithTimeout(60),
	)
	if err != nil {
		log.Fatalf("target parameters for router1 are incorrect: %s", err)
	}

	// Connect to the targets
	conn1, ctx1, err := xr.Connect(*router1)
	if err != nil {
		log.Fatalf("could not setup a client connection to %s, %v", router1.Host, err)
	}
	defer conn1.Close()

	// Get the BGP config from one of the devices
	id++
	ctx1, cancel := context.WithCancel(ctx1)
	defer cancel()
	c := make(chan os.Signal, 1)
	// If no signals are provided, all incoming signals will be relayed to c.
	// Otherwise, just the provided signals will. E.g.: signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()

	// Telemetry Subscription
	p := "lldp-dial-in-subs" // Change as needed, this should be configured on the device in advance. See README file for instructions
	// encoding GPB
	var e int64 = 2
	ch, ech, err := xr.GetSubscription(ctx1, conn1, p, id, e)
	if err != nil {
		log.Fatalf("could not setup Telemetry Subscription: %v\n", err)
	}

	go func() {
		select {
		case <-c:
			fmt.Printf("\nmanually cancelled the session to %v\n\n", router1.Host)
			cancel()
			return
		case <-ctx1.Done():
			// Timeout: "context deadline exceeded"
			err = ctx1.Err()
			fmt.Printf("\ngRPC session timed out after %v seconds: %v\n\n", router1.Timeout, err.Error())
			return
		case err = <-ech:
			// Session canceled: "context canceled"
			fmt.Printf("\ngRPC session to %v failed: %v\n\n", router1.Host, err.Error())
			return
		}
	}()

	fmt.Printf("\ntelemetry from %s\n\n", router1.Host)

	for tele := range ch {
		log.Printf("***** New message from %v ***** \n", router1.Host)
		message := new(telemetry.Telemetry)

		err = proto.Unmarshal(tele, message)
		if err != nil {
			log.Printf("Could not unmarshall the interface telemetry message for %v: %v\n", router1.Host, err)
		}
		for _, row := range message.GetDataGpb().GetRow() {
			// Get the message content

			content := row.GetContent()
			//fmt.Printf("\n\n%v\n\n\n", hex.Dump(content))
			nbr := new(lldp.LldpNeighbor)
			err = proto.Unmarshal(content, nbr)
			if err != nil {
				log.Fatalf("Could decode Content: %v\n", err)
			}
			for _, item := range nbr.GetLldpNeighbor() {
				deviceId := item.GetDeviceId()

				capabilities := item.GetEnabledCapabilities()
				chassisId := item.GetChassisId()
				// Print result:
				fmt.Printf("Device: %v\n Cappabilities: %v\n Chassis ID: %v\n\n", deviceId, capabilities, chassisId)

			}
		}
	}
}
