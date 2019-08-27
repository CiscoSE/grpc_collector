/*
Dial-in to specified router in specified subscription
*/

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/proto"
	xr "github.com/nleiva/xrgrpc"

	"github.com/nleiva/xrgrpc/proto/telemetry"
)

func main() {

	// Determine the ID for first the transaction.
	var id int64 = 1000

	// Manually specify target parameters.
	router1, err := xr.BuildRouter(
		xr.WithUsername("YOURUSER"),        // e.g. admin
		xr.WithPassword("YOUPASSWORD"),     // e.g. cisco123
		xr.WithHost("DEVICEIP:DEVICEPORT"), // e.g. 192.168.0.1:57344
		xr.WithTimeout(1000000),            // TODO, timeout shouldn't be required
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
	p := "Sub1" // Change as needed, this should be configured on the device in advance. See README file for instructions
	// encoding KVGPB
	var e int64 = 3
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

	fmt.Printf("\nConnected to %s\n\n", router1.Host)
	var namebuf bytes.Buffer
	for tele := range ch {
		log.Printf("***** New message from %v ***** \n", router1.Host)
		message := new(telemetry.Telemetry)

		err = proto.Unmarshal(tele, message)
		if err != nil {
			log.Printf("Could not unmarshall the interface telemetry message for %v: %v\n", router1.Host, err)
		}

		for _, gpbkv := range message.DataGpbkv {
			// Define file map
			var fields map[string]interface{}

			// Produce metadata tags
			var tags map[string]string

			// Top-level field may have measurement timestamp, if not use message timestamp
			measured := gpbkv.Timestamp
			if measured == 0 {
				measured = message.MsgTimestamp
			}

			timestamp := time.Unix(int64(measured/1000), int64(measured%1000)*1000000)

			// Populate tags and fields from toplevel GPBKV fields "keys" and "content"
			for _, field := range gpbkv.Fields {
				switch field.Name {
				case "keys":
					tags = make(map[string]string, len(field.Fields)+2)
					tags["Producer"] = message.GetNodeIdStr()
					tags["Target"] = message.GetSubscriptionIdStr()
					tags["EncodingPath"] = message.EncodingPath
					tags["TimeStamp"] = timestamp.String()
					for _, subfield := range field.Fields {
						parseGPBKVField(subfield, &namebuf, message.EncodingPath, timestamp, tags, nil)
					}
				case "content":
					fields = make(map[string]interface{}, len(field.Fields))
					for _, subfield := range field.Fields {
						parseGPBKVField(subfield, &namebuf, message.EncodingPath, timestamp, tags, fields)
					}
				default:
					log.Printf("I! Unexpected top-level MDT field: %s", field.Name)
				}
			}

			// Print measurement
			if len(fields) > 0 && len(tags) > 0 && len(message.EncodingPath) > 0 {

				log.Printf("\n**** New Telemetry message from %v ****", tags["Producer"])
				//log.Printf("Tags: %v", tags)
				//log.Printf("Fields: %v\n", fields)
				//log.Printf(telemetry.EncodingPath, fields, tags, timestamp)

			} else {
				fmt.Printf("I! Cisco MDT invalid field: encoding path or measurement empty")
			}
		}
	}

}

// Recursively parse GPBKV field structure into fields or tags
func parseGPBKVField(field *telemetry.TelemetryField, namebuf *bytes.Buffer,
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
		fmt.Printf("%v: %v\n", namebuf.String(), value)
	}

	for _, subfield := range field.Fields {
		parseGPBKVField(subfield, namebuf, path, timestamp, tags, fields)
	}

	namebuf.Truncate(namelen)
}
