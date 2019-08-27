# gNMI collector

A simple script to test gNMI telemetry.

## Installation and Usage

* Make sure to have [Go installed](https://golang.org/dl/)
* Install the dependencies using the following script [here](/install.sh)
* Add your device details and (optional) modify sensor paths at the end of the script. 
* Configure gRPC in your device. For example, in IOS-XR should look like this:

```
grpc
 port 57344
 address-family ipv4
!
```

* Build and run the application: 

```bash
cd $GOPATH/src/github.com/CiscoSE/grpc_collector/gnmi
go build
./gnmi
```
