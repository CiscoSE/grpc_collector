# Dial-in telemetry example

## Usage

This app shows telemetry messages from an XR device using cGBP and gRPC. Before executing this tool, make sure you have done the following:

### Create telemetry configuration for XR

```
grpc
 port 57344
 address-family ipv4
!
telemetry model-driven
 sensor-group lldp-dial-in-demo
  sensor-path Cisco-IOS-XR-ethernet-lldp-oper:lldp/nodes/node/neighbors/summaries/summary
exit
subscription lldp-dial-in-subs
  sensor-group-id lldp-dial-in-demo sample-interval 5000
commit

```

Note that if you use a different sensor path, you will need to update the code and the proto generated files might need to be updated too.

## Installation

* Make sure to have [Go installed](https://golang.org/dl/)
* Run the installation script located [here](/install.sh)
* Run the application: 

```bash
cd $GOPATH/github.com/CiscoSE/grpc_collector/dial_in
go build
./dial_in
```
