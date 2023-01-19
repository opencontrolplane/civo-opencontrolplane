package main

import (
	"context"

	"github.com/civo/civogo"
	"k8s.io/apimachinery/pkg/types"

	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) ListSSHKey(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.SSHKeyList, error) {
	client := ctx.Value("client").(*civogo.Client)

	sshKeys, err := client.ListSSHKeys()
	if err != nil {
		return nil, err
	}

	sshKeyList := &opencpspec.SSHKeyList{}
	for _, sshKey := range sshKeys {
		sshKeyList.Items = append(sshKeyList.Items, &opencpspec.SSHKey{
			Metadata: &metav1.ObjectMeta{
				Name:              sshKey.Name,
				UID:               types.UID(sshKey.ID),
				CreationTimestamp: metav1.Time{Time: sshKey.CreatedAt},
			},
			Spec: &opencpspec.SSHKeySpec{
				Publickey: sshKey.PublicKey,
			},
			Status: &opencpspec.SSHKeyStatus{
				Fingerprint: sshKey.Fingerprint,
				State:       "Active",
			},
		})
	}

	return sshKeyList, nil
}

func (s *Server) GetSSHKey(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.SSHKey, error) {
	client := ctx.Value("client").(*civogo.Client)

	sshKey, err := client.FindSSHKey(*option.Name)
	if err != nil {
		return nil, err
	}

	return &opencpspec.SSHKey{
		Metadata: &metav1.ObjectMeta{
			Name:              sshKey.Name,
			UID:               types.UID(sshKey.ID),
			CreationTimestamp: metav1.Time{Time: sshKey.CreatedAt},
		},
		Spec: &opencpspec.SSHKeySpec{
			Publickey: sshKey.PublicKey,
		},
		Status: &opencpspec.SSHKeyStatus{
			Fingerprint: sshKey.Fingerprint,
			State:       "Active",
		},
	}, nil
}

func (s *Server) CreateSSHKey(ctx context.Context, in *opencpspec.SSHKey) (*opencpspec.SSHKey, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Create the SSH key
	_, err := client.NewSSHKey(in.Metadata.Name, in.Spec.Publickey)
	if err != nil {
		return nil, err
	}

	// Get the SSH key
	sshKey, err := s.GetSSHKey(ctx, &opencpspec.FilterOptions{Name: &in.Metadata.Name})
	if err != nil {
		return nil, err
	}

	return sshKey, nil
}

func (s *Server) DeleteSSHKey(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.SSHKey, error) {
	client := ctx.Value("client").(*civogo.Client)

	sshKey, err := client.FindSSHKey(*option.Name)
	if err != nil {
		return nil, err
	}

	_, err = client.DeleteSSHKey(sshKey.ID)
	if err != nil {
		return nil, err
	}

	return &opencpspec.SSHKey{
		Metadata: &metav1.ObjectMeta{
			Name:              sshKey.Name,
			UID:               types.UID(sshKey.ID),
			CreationTimestamp: metav1.Time{Time: sshKey.CreatedAt},
		},
		Spec: &opencpspec.SSHKeySpec{
			Publickey: sshKey.PublicKey,
		},
		Status: &opencpspec.SSHKeyStatus{
			Fingerprint: sshKey.Fingerprint,
			State:       "Active",
		},
	}, nil
}
