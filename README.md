# gRPC Collector

Test telemetry collector using gRPC and KVGPB for XR, XE and NX devices


## Technical Challenge

Altough there are several collectors available for model driven telemetry, for cases when you just want to test what is coming back from the device, it can be difficult to find a solution that can just accept gRPC connection and print the telemetry messages in stdout. 
This app is based from the [Cisco MDT Telegraf Plugin](https://github.com/ios-xr/telegraf-plugin/tree/master/plugins/inputs/cisco_telemetry_mdt)


## Proposed Solution

The gRPC collector is able to open a port that devices can connect to via gRPC, using KV GPB as encoding without TLS.
This solution is aimed for testing purposes only. If you are looking for production grade collector contact Cisco or use an open source solution such as [Pipeline](https://github.com/cisco/bigmuddy-network-telemetry-pipeline) or [Telegraf Plugin](https://github.com/ios-xr/telegraf-plugin/tree/master/plugins/inputs/cisco_telemetry_mdt)


### Cisco Products Technologies/ Services

The app levegerage the following Cisco technologies

* [Cisco XE](https://www.cisco.com/c/en/us/products/ios-nx-os-software/ios-xe/index.html)
* [Cisco NX](https://www.cisco.com/c/en/us/products/ios-nx-os-software/nx-os/index.html)
* [Cisco XR](https://www.cisco.com/c/en/us/products/ios-nx-os-software/ios-xr-software/index.html)

## Authors

* Santiago Flores Kanter <sfloresk@cisco.com>


## Solution Components

* Go


## Usage

Once you have the solution installed, just configure the devices with model driven telemetry and point them to this collector.

### MDT Configuration example for XE

Enable netconf-yang processes
```
netconf-yang
```

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

Enable features

```
feature telemetry
feature nxapi
```


```

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
commit

```

TPA is needed in order to stream telemetry. Here is an example for vrf default

```
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

## Installation

* Make sure to have [Go installed](https://golang.org/dl/)
* Run the installation script located [here](./install.sh)
* Run the application: 

```bash
./bin/cisco_telemetry_mdt 
```

## Documentation

No extra documentation at this moment


## License

Provided under Cisco Sample Code License, for details see [LICENSE](./LICENSE.md)

## Code of Conduct

Our code of conduct is available [here](./CODE_OF_CONDUCT.md)

## Contributing

See our contributing guidelines [here](./CONTRIBUTING.md)
