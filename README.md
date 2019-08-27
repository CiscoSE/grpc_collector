# gRPC Collector

A set of go scripts to test telemetry using gRPC

## Technical Challenge

Altough there are several collectors available for model driven telemetry and gNMI, for cases when you just want to test what is coming back from the device, it can be difficult to find a solution that can only accept or create a gRPC connection and print the telemetry messages in stdout. 
This app is based from the [Cisco MDT Telegraf Plugin](https://github.com/ios-xr/telegraf-plugin/tree/master/plugins/inputs/cisco_telemetry_mdt) and [Cisco gNMI Telegraf Plugin](https://github.com/ios-xr/telegraf-plugin/tree/master/plugins/inputs/cisco_telemetry_mdt)


## Solution

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


## Installation and Usage 

* Make sure to have [Go installed](https://golang.org/dl/)

Please see README.md page for each script.

1. [gNMI](./gnmi) 
2. [Cisco Model Driven Telemetry - Dial out with KV](./cisco_telemetry_mdt/dial_out)
3. [Cisco Model Driven Telemetry - Dial in with compact GPB](./cisco_telemetry_mdt/dial_in)
4. [Cisco Model Driven Telemetry - Dial in with KV GPB](./cisco_telemetry_mdt/dial_in_kv)

## Documentation

No extra documentation at this moment

## License

Provided under Cisco Sample Code License, for details see [LICENSE](./LICENSE.md)

## Code of Conduct

Our code of conduct is available [here](./CODE_OF_CONDUCT.md)

## Contributing

See our contributing guidelines [here](./CONTRIBUTING.md)
