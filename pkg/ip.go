package pkg

import (
	"context"

	"github.com/civo/civogo"

	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (s *Server) ListIp(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.IpList, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Get all IPs and return them
	allIps, err := client.ListIPs()
	if err != nil {
		return nil, err
	}

	// Convert the IPs to the opencp format
	var ips []*opencpspec.Ip
	for _, ip := range allIps.Items {
		ips = append(ips, &opencpspec.Ip{
			Metadata: &metav1.ObjectMeta{
				Name: ip.Name,
				UID:  types.UID(ip.ID),
			},
			Spec: &opencpspec.IpSpec{
				Name: ip.Name,
			},
			Status: &opencpspec.IpStatus{
				Ip: ip.IP,
				Assignedto: &opencpspec.AssignedTo{
					Id:   ip.AssignedTo.ID,
					Type: ip.AssignedTo.Type,
					Name: ip.AssignedTo.Name,
				},
			},
		})
	}

	return &opencpspec.IpList{
		Items: ips,
	}, nil
}

func (s *Server) CreateIp(ctx context.Context, in *opencpspec.Ip) (*opencpspec.Ip, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Create the IP request
	ipRequest := &civogo.CreateIPRequest{
		Name:   in.Spec.Name,
		Region: client.Region,
	}

	// Create the IP
	ip, err := client.NewIP(ipRequest)
	if err != nil {
		return nil, err
	}

	// Get the IP from the GRPC server
	return s.GetIp(ctx, &opencpspec.FilterOptions{Name: &ip.Name})
}

func (s *Server) DeleteIp(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.Ip, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Get the IP
	ip, err := s.GetIp(ctx, option)
	if err != nil {
		return nil, err
	}

	// Delete the IP
	_, err = client.DeleteIP(string(ip.Metadata.UID))
	if err != nil {
		return nil, err
	}

	return ip, nil
}

func (s *Server) GetIp(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.Ip, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Get the IP
	ip, err := client.FindIP(*option.Name)
	if err != nil {
		return nil, err
	}

	return &opencpspec.Ip{
		Metadata: &metav1.ObjectMeta{
			Name: ip.Name,
			UID:  types.UID(ip.ID),
		},
		Spec: &opencpspec.IpSpec{
			Name: ip.Name,
		},
		Status: &opencpspec.IpStatus{
			Ip: ip.IP,
			Assignedto: &opencpspec.AssignedTo{
				Id:   ip.AssignedTo.ID,
				Type: ip.AssignedTo.Type,
				Name: ip.AssignedTo.Name,
			},
		},
	}, nil
}
