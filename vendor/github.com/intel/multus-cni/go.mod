module github.com/intel/multus-cni

go 1.12

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/containernetworking/cni v0.7.1
	github.com/containernetworking/plugins v0.8.2
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/k8snetworkplumbingwg/network-attachment-definition-client v0.0.0-20191025120722-4c57cd5732f3
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	github.com/vishvananda/netlink v0.0.0-20181108222139-023a6dafdcdf
	github.com/vishvananda/netns v0.0.0-20190625233234-7109fa855b0f // indirect
	golang.org/x/net v0.0.0-20190930134127-c5a3c61f89f3
	google.golang.org/grpc v1.23.0
	k8s.io/api v0.0.0-20181115043458-b799cb063522
	k8s.io/apimachinery v0.0.0-20181110190943-2a7c93004028
	k8s.io/client-go v0.0.0-20181115111358-9bea17718df8
	k8s.io/kubernetes v1.13.0
)
