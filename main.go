package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/civo/civogo"
	opencpspec "github.com/opencontrolplane/opencp-spec/v1alpha1"
	"google.golang.org/grpc"
)

type Server struct {
	opencpspec.LoginServer
	opencpspec.VirtualMachineServiceServer
	opencpspec.KubernetesClusterServiceServer
}

func (s *Server) Check(ctx context.Context, in *opencpspec.LoginRequest) (*opencpspec.LoginResponse, error) {

	regionCode := os.Getenv("CIVO_REGION")
	userAgent := &civogo.Component{
		Name:    "opencp.io",
		Version: "1.0.0",
	}
	client, err := civogo.NewClient(in.Token, regionCode)
	if err != nil {
		log.Println(err)
		return &opencpspec.LoginResponse{}, err
	}
	client.SetUserAgent(userAgent)

	result := client.GetAccountID()
	if result == "" {
		return &opencpspec.LoginResponse{Valid: false}, nil
	}

	return &opencpspec.LoginResponse{Valid: true}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 8080))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	
	opencpspec.RegisterLoginServer(s, &Server{})
	opencpspec.RegisterVirtualMachineServiceServer(s, &Server{})
	opencpspec.RegisterKubernetesClusterServiceServer(s, &Server{})

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
