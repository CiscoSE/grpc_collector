# Dial out gRPC Collector

Test telemetry collection using gRPC and KVGPB for XR, XE and NX devices

## Installation

* Make sure to have [Go installed](https://golang.org/dl/)
* Run the installation script located [here](./install.sh)
* Compile and run the application: 

```bash
cd $GOPATH/src/github.com/CiscoSE/grpc_collector/cisco_telemetry_mdt/dial_out
go build
./dial_out
```

## Usage

Once you have the solution installed, just configure the devices with model driven telemetry and point them to this collector.

### MDT Configuration example for XE

```
netconf-yang
netconf ssh
!
telemetry ietf subscription 0
 encoding encode-kvgpb
 filter xpath /memory-ios-xe-oper:memory-statistics/memory-statistic
 stream yang-push
 update-policy periodic 500
 receiver ip address <COLLECTOR_IP> 10000 protocol grpc-tcp
end

```

### MDT Configuration example for NX

```
feature telemetry
feature nxapi

config
telemetry
  destination-group 1
    ip address <COLLECTOR_IP> port 10000 protocol gRPC encoding GPB 
  sensor-group 1
    path sys/bgp/inst/dom-default depth 0
    path sys/mgmt-[eth1/1]/dbgIfIn
  sensor-group 2
    data-source NX-API
    path "show processes cpu" depth unbounded
    path "show processes memory physical" depth unbounded
    path "show system resources" depth unbounded
  subscription 1
    dst-grp 1
    snsr-grp 1 sample-interval 5000
    snsr-grp 2 sample-interval 5000
```

### MDT Configuration example for XR

```
config
telemetry model-driven
 destination-group grpc-collector
  address-family ipv4 <COLLECTOR_IP> port 10000
   encoding self-describing-gpb
   protocol grpc no-tls
  !
 sensor-group grpc-collector-sensor
  sensor-path Cisco-IOS-XR-nto-misc-oper:memory-summary/nodes/node/summary
  sensor-path Cisco-IOS-XR-infra-statsd-oper:infra-statistics/interfaces/interface/latest/generic-counters
 !
 subscription Sub1
  sensor-group-id grpc-collector-sensor sample-interval 5000
  destination-id grpc-collector
 !
!
tpa
 vrf default
  address-family ipv4
   default-route mgmt
   update-source dataports MgmtEth0/RP0/CPU0/0
  !
  address-family ipv6
   default-route mgmt
   update-source dataports MgmtEth0/RP0/CPU0/0
  !
 !
!
commit

```


