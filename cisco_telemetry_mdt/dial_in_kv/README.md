# Cisco MDT dial-in with KV GPB

A simple script for Cisco MDT dial-in telemetry using KV GPB as encoding. Tested only with IOS XR

## Installation and Usage

* Make sure to have [Go installed](https://golang.org/dl/)
* Install the dependencies using the [install](/install.sh) script
* Add your device details and (optional) modify subscription name hardcoded in the script. 
* Configure telemetry in your device. For example:

```
config
telemetry model-driven
 sensor-group grpc-collector-sensor
  sensor-path Cisco-IOS-XR-nto-misc-oper:memory-summary/nodes/node/summary
  sensor-path Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters
 !
 subscription Sub1
  destination-id grpc-collector
 !
!
commit

```

* Build and run the application: 

```bash
cd $GOPATH/src/github.com/CiscoSE/grpc_collector/cisco_telemetry_mdt/dial_in_kv
go build
./dial_in_kv
```

### MDT Configuration example for XR




