mkdir -p go/src
mkdir -p go/pkg
mkdir -p go/bin
cd go/
export GOPATH=$PWD
go get github.com/CiscoSE/grpc_collector
go install github.com/CiscoSE/grpc_collector/cisco_telemetry_mdt
