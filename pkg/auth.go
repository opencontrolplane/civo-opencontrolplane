package pkg

import (
	"context"
	"log"
	"os"

	"github.com/civo/civogo"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
)

var Version = "1.0.0"

func (s *Server) Check(ctx context.Context, in *opencpspec.LoginRequest) (*opencpspec.LoginResponse, error) {
	client := ctx.Value("client").(*civogo.Client)
	result := client.GetAccountID()
	if result == "" {
		return &opencpspec.LoginResponse{Valid: false}, nil
	}

	return &opencpspec.LoginResponse{Valid: true}, nil
}

// exampleAuthFunc is used by a middleware to authenticate requests
func AuthMiddlewareFunc(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	if token != "" {
		regionCode := os.Getenv("REGION")
		userAgent := &civogo.Component{
			Name:    "opencp.io",
			Version: Version,
		}
		client, err := civogo.NewClient(token, regionCode)
		if err != nil {
			log.Println(err)
			return ctx, err
		}
		client.SetUserAgent(userAgent)
		ctx = context.WithValue(ctx, "client", client)
	}

	return ctx, nil
}
