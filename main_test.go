package main

import (
	"context"
	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"github.com/civo/civo-opencp/pkg"
	"log"
	"net"
)

func FakeServer(ctx context.Context) (opencpspec.SSHKeyServiceClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	opencpspec.RegisterSSHKeyServiceServer(baseServer, &pkg.Server{})
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	client := opencpspec.NewSSHKeyServiceClient(conn)

	return client, closer
}
