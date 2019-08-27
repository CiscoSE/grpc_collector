package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// CiscoTelemetryGNMI plugin instance
type CiscoTelemetryGNMI struct {
	Addresses     []string
	Subscriptions []Subscription

	// Optional subscription configuration
	Encoding    string
	Origin      string
	Prefix      string
	Target      string
	UpdatesOnly bool

	// Cisco IOS XR credentials
	Username string
	Password string

	// Redial
	Redial time.Duration

	// GRPC TLS settings
	EnableTLS bool

	// Internal state

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Subscription for a GNMI client
type Subscription struct {
	Origin string
	Path   string

	// Subscription mode and interval
	SubscriptionMode  string
	SampleInterval    time.Duration
	HeartbeatInterval time.Duration
	SuppressRedundant bool
}

// Start the http listener service
func (c *CiscoTelemetryGNMI) Start() error {
	fmt.Printf("\nStarting GNMI Server\n")
	var err error
	var ctx context.Context
	var tlscfg *tls.Config
	var request *gnmi.SubscribeRequest

	ctx, c.cancel = context.WithCancel(context.Background())

	// Validate configuration
	if request, err = c.newSubscribeRequest(); err != nil {
		return err
	} else if c.Redial.Nanoseconds() <= 0 {
		return fmt.Errorf("redial duration must be positive")
	}

	// TODO: Should not be hardcoded!
	tlscfg = &tls.Config{
		InsecureSkipVerify: true,
	}

	if len(c.Username) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, "username", c.Username, "password", c.Password)
	}

	// Create a goroutine for each device, dial and subscribe
	fmt.Printf("\nCreating subscriptions for %v \n", c.Addresses)
	c.wg.Add(len(c.Addresses))
	for _, addr := range c.Addresses {
		fmt.Printf("\nStarting for collection for %v\n", addr)
		go func(address string) {
			defer c.wg.Done()
			for ctx.Err() == nil {
				if err := c.subscribeGNMI(ctx, address, tlscfg, request); err != nil && ctx.Err() == nil {
					log.Printf("Unexpected error: %v", err)
				}

				select {
				case <-ctx.Done():
				case <-time.After(c.Redial):
				}
			}
		}(addr)
	}
	ch := make(chan os.Signal, 1)
	select {
	case <-ch:
		fmt.Printf("\nManually cancelled the session\n")
		return nil
	}
}

// Create a new GNMI SubscribeRequest
func (c *CiscoTelemetryGNMI) newSubscribeRequest() (*gnmi.SubscribeRequest, error) {
	// Create subscription objects
	subscriptions := make([]*gnmi.Subscription, len(c.Subscriptions))
	for i, subscription := range c.Subscriptions {
		gnmiPath, err := parsePath(subscription.Origin, subscription.Path, "")
		if err != nil {
			return nil, err
		}
		mode, ok := gnmi.SubscriptionMode_value[strings.ToUpper(subscription.SubscriptionMode)]
		if !ok {
			return nil, fmt.Errorf("invalid subscription mode %s", subscription.SubscriptionMode)
		}
		subscriptions[i] = &gnmi.Subscription{
			Path:              gnmiPath,
			Mode:              gnmi.SubscriptionMode(mode),
			SampleInterval:    uint64(subscription.SampleInterval.Nanoseconds()),
			SuppressRedundant: subscription.SuppressRedundant,
			HeartbeatInterval: uint64(subscription.HeartbeatInterval.Nanoseconds()),
		}
	}

	// Construct subscribe request
	gnmiPath, err := parsePath(c.Origin, c.Prefix, c.Target)
	if err != nil {
		return nil, err
	}
	// TODO: add support for json and json_ietf
	if c.Encoding != "proto" {
		return nil, fmt.Errorf("unsupported encoding %s", c.Encoding)
	}

	return &gnmi.SubscribeRequest{
		Request: &gnmi.SubscribeRequest_Subscribe{
			Subscribe: &gnmi.SubscriptionList{
				Prefix:       gnmiPath,
				Mode:         gnmi.SubscriptionList_STREAM,
				Encoding:     gnmi.Encoding(gnmi.Encoding_value[strings.ToUpper(c.Encoding)]),
				Subscription: subscriptions,
				UpdatesOnly:  c.UpdatesOnly,
			},
		},
	}, nil
}

// SubscribeGNMI and extract telemetry data
func (c *CiscoTelemetryGNMI) subscribeGNMI(ctx context.Context, address string, tlscfg *tls.Config, request *gnmi.SubscribeRequest) error {
	var opt grpc.DialOption
	if tlscfg != nil {
		opt = grpc.WithTransportCredentials(credentials.NewTLS(tlscfg))
	} else {
		opt = grpc.WithInsecure()
	}

	client, err := grpc.DialContext(ctx, address, opt)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}
	defer client.Close()

	subscribeClient, err := gnmi.NewGNMIClient(client).Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup subscription: %v", err)
	}

	if err = subscribeClient.Send(request); err != nil {
		return fmt.Errorf("failed to send subscription request: %v", err)
	}

	log.Printf("Connection to GNMI device %s established", address)
	defer log.Printf("Connection to GNMI device %s closed", address)
	for ctx.Err() == nil {
		var reply *gnmi.SubscribeResponse
		if reply, err = subscribeClient.Recv(); err != nil {
			if err != io.EOF && ctx.Err() == nil {
				return fmt.Errorf("aborted GNMI subscription: %v", err)
			}
			break
		}

		c.handleSubscribeResponse(address, reply)
	}
	return nil
}

// HandleSubscribeResponse message from GNMI and parse contained telemetry data
func (c *CiscoTelemetryGNMI) handleSubscribeResponse(address string, reply *gnmi.SubscribeResponse) {
	// Check if response is a GNMI Update and if we have a prefix to derive the measurement name
	response, ok := reply.Response.(*gnmi.SubscribeResponse_Update)
	if !ok {
		return
	}

	var prefix, prefixAliasPath string

	timestamp := time.Unix(0, response.Update.Timestamp)
	prefixTags := make(map[string]string)

	if response.Update.Prefix != nil {
		prefix, prefixAliasPath = c.handlePath(response.Update.Prefix, prefixTags, "")
	}
	prefixTags["source"], _, _ = net.SplitHostPort(address)
	prefixTags["path"] = prefix

	// Parse individual Update message and create measurements
	fmt.Printf("\n\n###### New message from %v ######\n", prefixTags["source"])
	fmt.Printf("Timestamp: %v\n", timestamp)
	// Prepare tags from prefix
	tags := make(map[string]string, len(prefixTags))
	for key, val := range prefixTags {
		tags[key] = val
	}
	fmt.Printf("\n\n**** TAGS *****\n")
	for key, val := range tags {
		fmt.Printf("%v: %v\n", key, val)
	}

	fmt.Printf("\n\n**** Values *****\n")
	for _, update := range response.Update.Update {
		aliasPath, fields := c.handleTelemetryField(update, tags, prefix)
		// Inherent valid alias from prefix parsing
		if len(prefixAliasPath) > 0 && len(aliasPath) == 0 {
			aliasPath = prefixAliasPath
		}

		for key, val := range fields {
			if len(aliasPath) > 0 {
				key = key[len(aliasPath)+1:]
			}
			fmt.Printf("%v: %v\n", key, val)
		}

	}
}

// HandleTelemetryField and add it to a measurement
func (c *CiscoTelemetryGNMI) handleTelemetryField(update *gnmi.Update, tags map[string]string, prefix string) (string, map[string]interface{}) {
	path, aliasPath := c.handlePath(update.Path, tags, prefix)

	var value interface{}
	var jsondata []byte

	switch val := update.Val.Value.(type) {
	case *gnmi.TypedValue_AsciiVal:
		value = val.AsciiVal
	case *gnmi.TypedValue_BoolVal:
		value = val.BoolVal
	case *gnmi.TypedValue_BytesVal:
		value = val.BytesVal
	case *gnmi.TypedValue_DecimalVal:
		value = val.DecimalVal
	case *gnmi.TypedValue_FloatVal:
		value = val.FloatVal
	case *gnmi.TypedValue_IntVal:
		value = val.IntVal
	case *gnmi.TypedValue_StringVal:
		value = val.StringVal
	case *gnmi.TypedValue_UintVal:
		value = val.UintVal
	case *gnmi.TypedValue_JsonIetfVal:
		jsondata = val.JsonIetfVal
	case *gnmi.TypedValue_JsonVal:
		jsondata = val.JsonVal
	}

	name := strings.Replace(path, "-", "_", -1)
	fields := make(map[string]interface{})
	if value != nil {
		fields[name] = value
	} else if jsondata != nil {
		if err := json.Unmarshal(jsondata, &value); err != nil {
			log.Printf("failed to parse JSON value: %v", err)
		}
	}
	return aliasPath, fields
}

// Parse path to path-buffer and tag-field
func (c *CiscoTelemetryGNMI) handlePath(path *gnmi.Path, tags map[string]string, prefix string) (string, string) {
	var aliasPath string
	builder := bytes.NewBufferString(prefix)

	// Prefix with origin
	if len(path.Origin) > 0 {
		builder.WriteString(path.Origin)
		builder.WriteRune(':')
	}

	// Parse generic keys from prefix
	for _, elem := range path.Elem {
		builder.WriteRune('/')
		builder.WriteString(elem.Name)
		name := builder.String()

		if tags != nil {
			for key, val := range elem.Key {
				key = strings.Replace(key, "-", "_", -1)

				// Use short-form of key if possible
				if _, exists := tags[key]; exists {
					tags[name+"/"+key] = val
				} else {
					tags[key] = val
				}

			}
		}
	}

	return builder.String(), aliasPath
}

//ParsePath from XPath-like string to GNMI path structure
func parsePath(origin string, path string, target string) (*gnmi.Path, error) {
	var err error
	gnmiPath := gnmi.Path{Origin: origin, Target: target}

	if len(path) > 0 && path[0] != '/' {
		return nil, fmt.Errorf("path does not start with a '/': %s", path)
	}

	elem := &gnmi.PathElem{}
	start, name, value, end := 0, -1, -1, -1

	path = path + "/"

	for i := 0; i < len(path); i++ {
		if path[i] == '[' {
			if name >= 0 {
				break
			}
			if end < 0 {
				end = i
				elem.Key = make(map[string]string)
			}
			name = i + 1
		} else if path[i] == '=' {
			if name <= 0 || value >= 0 {
				break
			}
			value = i + 1
		} else if path[i] == ']' {
			if name <= 0 || value <= name {
				break
			}
			elem.Key[path[name:value-1]] = strings.Trim(path[value:i], "'\"")
			name, value = -1, -1
		} else if path[i] == '/' {
			if name < 0 {
				if end < 0 {
					end = i
				}

				if end > start {
					elem.Name = path[start:end]
					gnmiPath.Elem = append(gnmiPath.Elem, elem)
					gnmiPath.Element = append(gnmiPath.Element, path[start:i])
				}

				start, name, value, end = i+1, -1, -1, -1
				elem = &gnmi.PathElem{}
			}
		}
	}

	if name >= 0 || value >= 0 {
		err = fmt.Errorf("Invalid GNMI path: %s", path)
	}

	if err != nil {
		return nil, err
	}

	return &gnmiPath, nil
}

// Stop listener and cleanup
func (c *CiscoTelemetryGNMI) Stop() {
	c.cancel()
	c.wg.Wait()
}

func main() {
	subscriptions := []Subscription{}
	addresses := []string{}

	// Define redial and sample interval.
	sampleInterval := 10 * time.Second
	redialInterval := 10 * time.Second

	// Sample subscription for interface counters
	subs := Subscription{
		Origin:           "openconfig-interfaces",
		Path:             "/interfaces/interface/state/counters",
		SubscriptionMode: "sample",
		SampleInterval:   sampleInterval,
	}

	// Add desired subscriptions to this slide
	subscriptions = append(subscriptions, subs)

	// Add the addresses that you would like to dial
	addresses = append(addresses, "YOURDEVICE:YOURGRPCPORT")

	// Collection parameters, including credentials. TLS not verified by default (hardcoded)
	gnmiCollector := CiscoTelemetryGNMI{
		Encoding:      "proto",
		Redial:        redialInterval,
		Username:      "YOURUSER",
		Password:      "YOURPASSWORD",
		Addresses:     addresses,
		Subscriptions: subscriptions,
	}

	err := gnmiCollector.Start()
	if err != nil {
		log.Printf("Error starting gnmi Collector: %v\n", err)
	}
}
