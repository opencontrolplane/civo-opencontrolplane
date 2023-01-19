package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logger := logrus.WithFields(logrus.Fields{})

	opts := []grpc_logrus.Option{
		grpc_logrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ns", duration.Nanoseconds()
		}),
	}

	grpc_logrus.ReplaceGrpcLogger(logger)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 8080))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc_middleware.WithStreamServerChain(
			grpc_auth.StreamServerInterceptor(AuthMiddlewareFunc),
		),
		grpc_middleware.WithUnaryServerChain(
			grpc_auth.UnaryServerInterceptor(AuthMiddlewareFunc),
			grpc_logrus.UnaryServerInterceptor(logger, opts...),
		),
	)

	opencpspec.RegisterLoginServer(grpcServer, &Server{})
	opencpspec.RegisterVirtualMachineServiceServer(grpcServer, &Server{})
	opencpspec.RegisterKubernetesClusterServiceServer(grpcServer, &Server{})
	opencpspec.RegisterNamespaceServiceServer(grpcServer, &Server{})
	opencpspec.RegisterDomainServiceServer(grpcServer, &Server{})
	opencpspec.RegisterSSHKeyServiceServer(grpcServer, &Server{})
	opencpspec.RegisterFirewallServiceServer(grpcServer, &Server{})

	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
