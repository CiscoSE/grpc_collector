mkdir -p go/src
mkdir go/pkg
mkdir go/bin
cd go/
export GOPATH=$PWD
go get github.com/CiscoSE/grpc_collector
go get github.com/nleiva/xrgrpc
go get github.com/golang/protobuf/proto
go get golang.org/x/net/context
go get google.golang.org/grpc
go get google.golang.org/grpc/peer
go get github.com/openconfig/gnmi

