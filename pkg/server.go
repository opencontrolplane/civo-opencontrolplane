package pkg

import (
	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
)

type Server struct {
	opencpspec.LoginServer
	opencpspec.VirtualMachineServiceServer
	opencpspec.KubernetesClusterServiceServer
	opencpspec.NamespaceServiceServer
	opencpspec.DomainServiceServer
	opencpspec.SSHKeyServiceServer
	opencpspec.FirewallServiceServer
}