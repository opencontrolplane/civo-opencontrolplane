package main

import (
	"context"
	"strconv"

	"github.com/civo/civogo"
	"k8s.io/apimachinery/pkg/types"

	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) ListVirtualMachine(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.VirtualMachineList, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Get all the virtual machines
	allvm, err := client.ListAllInstances()
	if err != nil {
		return nil, err
	}

	// Get all the networks
	network, err := s.ListNamespace(ctx, nil)
	if err != nil {
		return nil, err
	}

	// convert the virtual machines to the opencp format
	vms := []*opencpspec.VirtualMachine{}
	for _, vm := range allvm {
		// find the network
		var networkName string
		for _, net := range network.Items {
			if net.Metadata.UID == types.UID(vm.NetworkID) {
				networkName = net.Metadata.Name
			}
		}

		vms = append(vms, &opencpspec.VirtualMachine{
			Metadata: &metav1.ObjectMeta{
				Name:              vm.Hostname,
				Namespace:         networkName,
				UID:               types.UID(vm.ID),
				CreationTimestamp: metav1.Time{Time: vm.CreatedAt},
			},
			Spec: &opencpspec.VirtualMachineSpec{
				Size:     vm.Size,
				Firewall: vm.FirewallID,
				Ipv4:     false,
				Ipv6:     false,
				Image:    vm.SourceID,
				Auth: &opencpspec.VirtualMachineAuth{
					User:   vm.InitialUser,
					SshKey: vm.SSHKeyID,
				},
				Tags:       vm.Tags,
				UserScript: vm.Script,
			},
			Status: &opencpspec.VirtualMachineStatus{
				PrivateIP: vm.PrivateIP,
				PublicIP:  vm.PublicIP,
				State:     vm.Status,
			},
		})
	}

	// check if the filter is empty
	if option != nil {
		// check if namespace is set
		if option.Namespace != nil && *option.Namespace != "" {
			// filter the virtual machines by namespace
			filteredvms := []*opencpspec.VirtualMachine{}
			for _, vm := range vms {
				if vm.Metadata.Namespace == *option.Namespace {
					filteredvms = append(filteredvms, vm)
				}
			}

			return &opencpspec.VirtualMachineList{
				Items: filteredvms,
			}, nil
		}
	}

	return &opencpspec.VirtualMachineList{
		Items: vms,
	}, nil
}

func (s *Server) CreateVirtualMachine(ctx context.Context, in *opencpspec.VirtualMachine) (*opencpspec.VirtualMachine, error) {
	client := ctx.Value("client").(*civogo.Client)

	// TODO move this to a GRPC util function
	getDiskImage, err := client.FindDiskImage(in.Spec.Image)
	if err != nil {
		return nil, err
	}

	// Get the network
	network, err := s.GetNamespace(ctx, &opencpspec.FilterOptions{Name: &in.Metadata.Namespace})
	if err != nil {
		return nil, err
	}

	// Create the VM
	vm := &civogo.InstanceConfig{
		Hostname:         in.Metadata.Name,
		ReverseDNS:       in.Metadata.Name,
		Size:             in.Spec.Size,
		Region:           client.Region,
		PublicIPRequired: strconv.FormatBool(in.Spec.Ipv4),
		NetworkID:        string(network.Metadata.UID),
		TemplateID:       getDiskImage.ID,
		Script:           in.Spec.UserScript,
		Tags:             in.Spec.Tags,
	}

	// Check the firewall
	if in.Spec.Firewall != "" {
		getFirewall, err := client.FindFirewall(in.Spec.Firewall)
		if err != nil {
			return nil, err
		}
		vm.FirewallID = getFirewall.ID
	}

	if in.Spec.Auth.User != "" {
		vm.InitialUser = in.Spec.Auth.User
	}

	if in.Spec.Auth.SshKey != "" {
		vm.SSHKeyID = in.Spec.Auth.SshKey
	}

	// Create the cluster
	instance, err := client.CreateInstance(vm)
	if err != nil {
		return nil, err
	}

	// Get the virtual machine
	virtualMachine, err := s.GetVirtualMachine(ctx, &opencpspec.FilterOptions{Id: &instance.ID})
	if err != nil {
		return nil, err
	}

	return virtualMachine, nil
}

func (s *Server) GetVirtualMachine(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.VirtualMachine, error) {
	client := ctx.Value("client").(*civogo.Client)

	// check the options to see wish value to use
	// TODO do a better check, put this in a util function
	var filter string
	switch {
	case option.Id != nil:
		filter = *option.Id
	case option.Name != nil:
		filter = *option.Name
	}

	// Get the virtual machine
	vm, err := client.FindInstance(filter)
	if err != nil {
		return nil, err
	}

	// Get the network if the namespace is set in the options
	var networkName string
	if option.Namespace != nil && *option.Namespace != "" {
		network, err := s.GetNamespace(ctx, &opencpspec.FilterOptions{Name: option.Namespace})
		if err != nil {
			return nil, err
		}

		// filter the virtual machine by namespace
		if vm.NetworkID != string(network.Metadata.UID) {
			return nil, nil
		}

		// set the network name
		networkName = network.Metadata.Name
	}

	return &opencpspec.VirtualMachine{
		Metadata: &metav1.ObjectMeta{
			Name:              vm.Hostname,
			Namespace:         networkName,
			UID:               types.UID(vm.ID),
			CreationTimestamp: metav1.Time{Time: vm.CreatedAt},
		},
		Spec: &opencpspec.VirtualMachineSpec{
			Size:     vm.Size,
			Firewall: vm.FirewallID,
			Ipv4:     false,
			Ipv6:     false,
			Image:    vm.SourceID,
			Auth: &opencpspec.VirtualMachineAuth{
				User:   vm.InitialUser,
				SshKey: vm.SSHKeyID,
			},
			Tags:       vm.Tags,
			UserScript: vm.Script,
		},
		Status: &opencpspec.VirtualMachineStatus{
			PrivateIP: vm.PrivateIP,
			PublicIP:  vm.PublicIP,
			State:     vm.Status,
		},
	}, nil
}

func (s *Server) DeleteVirtualMachine(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.VirtualMachine, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Get the virtual machine
	virtualMachine, err := s.GetVirtualMachine(ctx, option)
	if err != nil {
		return nil, err
	}

	// Delete the virtual machine
	if virtualMachine != nil {
		_, err := client.DeleteInstance(string(virtualMachine.Metadata.UID))
		if err != nil {
			return nil, err
		}
	}

	return virtualMachine, nil
}

// UpdateVirtualMachine(context.Context, *VirtualMachine) (*VirtualMachine, error)
