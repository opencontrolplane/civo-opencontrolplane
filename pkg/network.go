package pkg

import (
	"context"
	"log"

	"github.com/civo/civogo"

	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ListNamespace returns a list of all the namespaces
func (s *Server) ListNamespace(ctx context.Context, in *opencpspec.FilterOptions) (*opencpspec.NamespaceList, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Get all the networks again and return them
	allNetwork, err := client.ListNetworks()
	if err != nil {
		log.Println(err)
	}

	// Convert the networks to the opencp format
	var networks []*opencpspec.Namespace
	for _, network := range allNetwork {
		networks = append(networks, &opencpspec.Namespace{
			Kind:       "Namespace",
			ApiVersion: "v1",
			Metadata: &metav1.ObjectMeta{
				Name: network.Label,
				UID:  types.UID(network.ID),
			},
			Spec: &corev1.NamespaceSpec{
				Finalizers: []corev1.FinalizerName{},
			},
			Status: &corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		})
	}

	// Return the list of networks
	return &opencpspec.NamespaceList{
		Kind:       "NamespaceList",
		ApiVersion: "v1",
		Items:      networks,
	}, nil
}

// CreateNamespace creates a new namespace
func (s *Server) CreateNamespace(ctx context.Context, in *opencpspec.Namespace) (*opencpspec.Namespace, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Create the network
	networkResult, err := client.NewNetwork(in.Metadata.Name)
	if err != nil {
		log.Println(err)
	}

	// Get the network
	network, err := client.GetNetwork(networkResult.ID)
	if err != nil {
		log.Println(err)
	}

	// Convert the network to the opencp format
	networks := &opencpspec.Namespace{
		Kind:       "Namespace",
		ApiVersion: "v1",
		Metadata: &metav1.ObjectMeta{
			Name: network.Label,
			UID:  types.UID(network.ID),
		},
		Spec: &corev1.NamespaceSpec{
			Finalizers: []corev1.FinalizerName{},
		},
		Status: &corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}

	return networks, nil
}

func (s *Server) GetNamespace(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.Namespace, error) {
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

	// Get all the networks again and return them
	network, err := client.FindNetwork(filter)
	if err != nil {
		log.Println(err)
	}

	// Convert the networks to the opencp format
	var networks *opencpspec.Namespace
	if network != nil {
		networks = &opencpspec.Namespace{
			Kind:       "Namespace",
			ApiVersion: "v1",
			Metadata: &metav1.ObjectMeta{
				Name: network.Label,
				UID:  types.UID(network.ID),
			},
			Spec: &corev1.NamespaceSpec{
				Finalizers: []corev1.FinalizerName{},
			},
			Status: &corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		}
	}

	return networks, nil
}

func (s *Server) DeleteNamespace(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.Namespace, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Get all the networks again and return them
	network, err := client.FindNetwork(*option.Id)
	if err != nil {
		log.Println(err)
	}

	if network != nil {
		_, err = client.DeleteNetwork(network.ID)
		if err != nil {
			log.Println(err)
		}
	}

	// Convert the networks to the opencp format
	var networks *opencpspec.Namespace
	if network != nil {
		networks = &opencpspec.Namespace{
			Kind:       "Namespace",
			ApiVersion: "v1",
			Metadata: &metav1.ObjectMeta{
				Name: network.Label,
				UID:  types.UID(network.ID),
			},
			Spec: &corev1.NamespaceSpec{
				Finalizers: []corev1.FinalizerName{},
			},
			Status: &corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		}
	}

	return networks, nil
}

// UpdateNamespace(context.Context, *Namespace) (*Namespace, error)
//
