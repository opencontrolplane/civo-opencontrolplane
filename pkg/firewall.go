package pkg

import (
	"context"
	"strconv"

	"github.com/civo/civogo"
	"k8s.io/apimachinery/pkg/types"

	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) ListFirewall(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.FirewallList, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Get all the firewall rules again and return them
	firewall, err := client.ListFirewalls()
	if err != nil {
		return nil, err
	}

	// Get all the networks
	network, err := s.ListNamespace(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Convert the firewall to the opencp spec
	firewalls := []*opencpspec.Firewall{}
	for _, fw := range firewall {
		// find the network
		var networkName string
		for _, net := range network.Items {
			if net.Metadata.UID == types.UID(fw.NetworkID) {
				networkName = net.Metadata.Name
			}
		}

		fwstatus := "Ready"
		if fw.ClusterCount == 0 && fw.InstanceCount == 0 && fw.LoadBalancerCount == 0 {
			fwstatus = "Inactive"
		}

		// Loop through the rules
		ingressRules := []*opencpspec.FirewallRules{}
		egressRules := []*opencpspec.FirewallRules{}

		for _, rule := range fw.Rules {
			if rule.Direction == "ingress" {
				ingressRules = append(ingressRules, &opencpspec.FirewallRules{
					Source:   rule.Cidr,
					Label:    rule.Label,
					Ports:    rule.Ports,
					Protocol: rule.Protocol,
					Action:   rule.Action,
				})
			} else {
				egressRules = append(egressRules, &opencpspec.FirewallRules{
					Source:   rule.Cidr,
					Label:    rule.Label,
					Ports:    rule.Ports,
					Protocol: rule.Protocol,
					Action:   rule.Action,
				})
			}
		}

		firewalls = append(firewalls, &opencpspec.Firewall{
			Metadata: &metav1.ObjectMeta{
				Name:      fw.Name,
				UID:       types.UID(fw.ID),
				Namespace: networkName,
			},
			Spec: &opencpspec.FirewallSpec{
				Ingress: ingressRules,
				Egress:  egressRules,
			},
			Status: &opencpspec.FirewallStatus{
				State:      fwstatus,
				Totalrules: strconv.Itoa(fw.RulesCount),
			},
		})
	}

	// check if the filter is empty
	if option != nil {
		// check if namespace is set
		if option.Namespace != nil && *option.Namespace != "" {
			// filter the virtual machines by namespace
			filteredFw := []*opencpspec.Firewall{}
			for _, fw := range firewalls {
				if fw.Metadata.Namespace == *option.Namespace {
					filteredFw = append(filteredFw, fw)
				}
			}

			return &opencpspec.FirewallList{
				Items: filteredFw,
			}, nil
		}
	}

	return &opencpspec.FirewallList{
		Items: firewalls,
	}, nil
}

func (s *Server) GetFirewall(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.Firewall, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	fw, err := client.FindFirewall(*option.Name)
	if err != nil {
		return nil, err
	}

	// find the network and filter the firewall
	var networkName string
	if option.Namespace != nil && *option.Namespace != "" {
		network, err := s.GetNamespace(ctx, &opencpspec.FilterOptions{Name: option.Namespace})
		if err != nil {
			return nil, err
		}

		// filter the firewall by network
		if fw.NetworkID != string(network.Metadata.UID) {
			return nil, nil
		}

		// set the network name
		networkName = network.Metadata.Name
	}

	// convert firewall to opencp spec
	fwstatus := "Ready"
	if fw.ClusterCount == 0 && fw.InstanceCount == 0 && fw.LoadBalancerCount == 0 {
		fwstatus = "Inactive"
	}

	// Loop through the rules
	ingressRules := []*opencpspec.FirewallRules{}
	egressRules := []*opencpspec.FirewallRules{}

	for _, rule := range fw.Rules {
		if rule.Direction == "ingress" {
			ingressRules = append(ingressRules, &opencpspec.FirewallRules{
				Source:   rule.Cidr,
				Label:    rule.Label,
				Ports:    rule.Ports,
				Protocol: rule.Protocol,
				Action:   rule.Action,
			})
		} else {
			egressRules = append(egressRules, &opencpspec.FirewallRules{
				Source:   rule.Cidr,
				Label:    rule.Label,
				Ports:    rule.Ports,
				Protocol: rule.Protocol,
				Action:   rule.Action,
			})
		}
	}

	return &opencpspec.Firewall{
		Metadata: &metav1.ObjectMeta{
			Name:      fw.Name,
			UID:       types.UID(fw.ID),
			Namespace: networkName,
		},
		Spec: &opencpspec.FirewallSpec{
			Ingress: ingressRules,
			Egress:  egressRules,
		},
		Status: &opencpspec.FirewallStatus{
			State:      fwstatus,
			Totalrules: strconv.Itoa(fw.RulesCount),
		},
	}, nil
}

func (s *Server) CreateFirewall(ctx context.Context, in *opencpspec.Firewall) (*opencpspec.Firewall, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// get the network
	network, err := s.GetNamespace(ctx, &opencpspec.FilterOptions{Name: &in.Metadata.Namespace})
	if err != nil {
		return nil, err
	}

	// create the firewall config
	createDefaultRules := false
	fwConfig := &civogo.FirewallConfig{
		Name:        in.Metadata.Name,
		Region:      client.Region,
		NetworkID:   string(network.Metadata.UID),
		CreateRules: &createDefaultRules,
	}

	// create the firewall rules
	for _, rule := range in.Spec.Ingress {
		fwConfig.Rules = append(fwConfig.Rules, civogo.FirewallRule{
			Direction: "ingress",
			Cidr:      rule.Source,
			Label:     rule.Label,
			Ports:     rule.Ports,
			Protocol:  rule.Protocol,
			Action:    rule.Action,
		})
	}

	for _, rule := range in.Spec.Egress {
		fwConfig.Rules = append(fwConfig.Rules, civogo.FirewallRule{
			Direction: "egress",
			Cidr:      rule.Source,
			Label:     rule.Label,
			Ports:     rule.Ports,
			Protocol:  rule.Protocol,
			Action:    rule.Action,
		})
	}

	// create the firewall
	fw, err := client.NewFirewall(fwConfig)
	if err != nil {
		return nil, err
	}

	// get the firewall and return
	return s.GetFirewall(ctx, &opencpspec.FilterOptions{Name: &fw.Name})
}

func (s *Server) DeleteFirewall(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.Firewall, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Get the firewall
	fw, err := s.GetFirewall(ctx, option)
	if err != nil {
		return nil, err
	}

	// Delete the firewall
	if fw != nil {
		_, err = client.DeleteFirewall(string(fw.Metadata.UID))
		if err != nil {
			return nil, err
		}
	}

	return fw, nil
}
