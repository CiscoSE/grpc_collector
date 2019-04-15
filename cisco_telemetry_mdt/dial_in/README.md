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

Note that if you use a different sensor path, you will need to update the code and possible, the proto generated files


### Update the certificate

Copy the certificate from the router to the [ems.pem file](./ems.pem). To get get the cert in the XR device, you can use the following command:

```
bash cat /misc/config/grpc/ems.pem
```

You will see something like this:

```
-----BEGIN CERTIFICATE-----
test123test123test123test123test123test123test123test123test1233
MQswCQYDVQQIEwJDQTERMA8GA1UEBxMIU2FuIEpvc2UxFzAVBgNVBAkTDjM3MDAg
test123test123test123test123test123test123test123test123test1233
cywgSW5jLjEMMAoGA1UECxMDQ1NHMRYwFAYDVQQDEw1lbXMuY2lzY28uY29tMRQw
EgYDVQQFEwtGT0MyMTAzUjA5RTAeFw0xOTA0MDQyMDM0MTZaFw0zOTA0MDQyMDM0
MTZaMIGwMQswCQYDVQQGEwJVUzELMAkGA1UECBMCQ0ExETAPBgNVBAcTCFNhbiBK
b3NlMRcwFQYDVQQJEw4zNzAwIENpc2NvIFdheTEOMAwGA1UEERMFOTUxMzQxHDAa
BgNVBAoTE0Npc2NvIFN5c3RlbXMsIEluYy4xDDAKBgNVBAsTA0NTRzEWMBQGA1UE
AxMNZW1zLmNpc2NvLmNvbTEUMBIGA1UEBRMLRk9DMjEwM1IwOUUwggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQDoRQpwDCgUZ4LmGxfDYdQPfqVVr68pp4cd
HOYbOd2YMlYaHq+MIKiH0yzaoF8JJDY0dQSx246wPIAMuF5YI3oKmNmrbrocILKt
s0KkpapyzJ53aOc/j7/mjv+0Yj2z4oGN/hYof6bc2sQnxI9RDwb/rCfams17K7i4
VMedFkXvFCtgyFSMsArX7n5rnGxdXHA5zBLkgmmg3LQuweSjIS8QUi6BOLkrOwQv
test123test123test123test123test123test123test123test123test1233
9wrp1jbTYkI7/v+QmSQu24fyUE5PJ1p5rw5oX78TL3zjVWiDg/q5AgMBAAGjMjAw
MA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFNuFNPZM+mznC1YY/CF7ndG9JWYU
MA0GCSqGSIb3DQEBDQUAA4IBAQA8JYxdWKhKRjKtIZ146t1zUSugzSdnRQQagnuo
ecjccLSBSO6nz1Eipbed9X0e7Yywrj6MkEWVozIO6+5KkHwuGsW2wp3PBX/b7Y8Z
a5QMIhYJEUKwtbZ5821+ppldw8uIuurL/cEoUjB55ZL18hacDYtPVonm3j70azFJ
MlCXgnhIK6x5erARZ7LmOORg5BVcPUT8m4rKk5eMFU0s/egcWDljLmn7zJqZ/PPp
test123test123test123test123test123test123test123test123test1233
lGClNOl8YLZCnDlYoPspmQy14ttMRQlqNobEObbmeObKFBiL
-----END CERTIFICATE-----
```

Copy the entire certificate to the [ems.pem file](./ems.pem) file.

## Installation

* Make sure to have [Go installed](https://golang.org/dl/)
* Run the installation script located [here](/install.sh)
* Run the application: 

```bash
cd $GOPATH/github.com/CiscoSE/grpc_collector/dial_in
go build
./dial_in
```

## Documentation

No extra documentation at this moment


## License

Provided under Cisco Sample Code License, for details see [LICENSE](/LICENSE.md)

## Code of Conduct

Our code of conduct is available [here](/CODE_OF_CONDUCT.md)

## Contributing

See our contributing guidelines [here](/CONTRIBUTING.md)
