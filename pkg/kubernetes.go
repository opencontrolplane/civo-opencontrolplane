package pkg

import (
	"context"

	"github.com/civo/civogo"
	"k8s.io/apimachinery/pkg/types"

	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) CreateKubernetesCluster(ctx context.Context, in *opencpspec.KubernetesCluster) (*opencpspec.KubernetesCluster, error) {
	client := ctx.Value("client").(*civogo.Client)

	// get the network
	network, err := s.GetNamespace(ctx, &opencpspec.FilterOptions{Name: &in.Metadata.Namespace})
	if err != nil {
		return nil, err
	}

	// convert the pools
	pools := []civogo.KubernetesClusterPoolConfig{}
	for _, pool := range in.Spec.Pools {
		pools = append(pools, civogo.KubernetesClusterPoolConfig{
			ID:    pool.Id,
			Size:  pool.Size,
			Count: int(pool.Count),
		})
	}

	// Create a kubernetes cluster config
	k8sConfig := &civogo.KubernetesClusterConfig{
		Name:              in.Metadata.Name,
		Region:            client.Region,
		KubernetesVersion: in.Spec.Version,
		NetworkID:         string(network.Metadata.UID),
		Pools:             pools,
		CNIPlugin:         in.Spec.Cniplugin,
	}

	if in.Spec.Clustertype != "" {
		k8sConfig.ClusterType = in.Spec.Clustertype
	}

	// Check if the incoming cluster have firewall
	if in.Spec.Firewall != "" {
		// Get the firewall object
		firewall, err := s.GetFirewall(ctx, &opencpspec.FilterOptions{Name: &in.Spec.Firewall})
		if err != nil {
			return nil, err
		}
		k8sConfig.InstanceFirewall = string(firewall.Metadata.UID)
	}

	// create the kubernetes cluster
	kubernetesCluster, err := client.NewKubernetesClusters(k8sConfig)
	if err != nil {
		return nil, err
	}

	// Get the latest version of the kubernetes cluster
	kubernetesClusterLast, err := s.GetKubernetesCluster(ctx, &opencpspec.FilterOptions{Id: &kubernetesCluster.ID})
	if err != nil {
		return nil, err
	}

	return kubernetesClusterLast, nil
}

func (s *Server) GetKubernetesCluster(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.KubernetesCluster, error) {
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

	// Get the kubernetes cluster
	k8s, err := client.FindKubernetesCluster(filter)
	if err != nil {
		return nil, err
	}

	var networkName string
	if option.Namespace != nil && *option.Namespace != "" {
		network, err := s.GetNamespace(ctx, &opencpspec.FilterOptions{Name: option.Namespace})
		if err != nil {
			return nil, err
		}

		// filter the virtual machine by namespace
		if k8s.NetworkID != string(network.Metadata.UID) {
			return nil, nil
		}

		// set the network name
		networkName = network.Metadata.Name
	}

	// Get the firewall
	firewall, err := s.GetFirewall(ctx, &opencpspec.FilterOptions{Name: &k8s.FirewallID})
	if err != nil {
		return nil, err
	}

	// convert the pools
	pools := []*opencpspec.KubernetesClusterPool{}
	for _, pool := range k8s.Pools {
		pools = append(pools, &opencpspec.KubernetesClusterPool{
			Id:    pool.ID,
			Size:  pool.Size,
			Count: int32(pool.Count),
		})
	}

	return &opencpspec.KubernetesCluster{
		Metadata: &metav1.ObjectMeta{
			Name:              k8s.Name,
			Namespace:         networkName,
			UID:               types.UID(k8s.ID),
			CreationTimestamp: metav1.NewTime(k8s.CreatedAt),
		},
		Spec: &opencpspec.KubernetesClusterSpec{
			Pools:       pools,
			Version:     k8s.Version,
			Firewall:    firewall.Metadata.Name,
			Cniplugin:   k8s.CNIPlugin,
			Clustertype: k8s.ClusterType,
			Kubeconfig:  k8s.KubeConfig,
		},
		Status: &opencpspec.KubernetesClusterStatus{
			State:    k8s.Status,
			Endpoint: k8s.APIEndPoint,
			Publicip: k8s.MasterIP,
		},
	}, nil
}

func (s *Server) ListKubernetesCluster(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.KubernetesClusterList, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Get all the kubernetes clusters
	allk8s, err := client.ListKubernetesClusters()
	if err != nil {
		return nil, err
	}

	// Get all the networks
	network, err := s.ListNamespace(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Get all firewall from Civo
	firewall, err := s.ListFirewall(ctx, option)
	if err != nil {
		return nil, err
	}

	// convert the kubernetes clusters to the opencp format
	kubernetesCluster := []*opencpspec.KubernetesCluster{}
	for _, k8s := range allk8s.Items {
		// find the network
		var networkName string
		for _, net := range network.Items {
			if net.Metadata.UID == types.UID(k8s.NetworkID) {
				networkName = net.Metadata.Name
			}
		}

		// Get the right firewall
		var firewallName string
		for _, fw := range firewall.Items {
			if fw.Metadata.UID == types.UID(k8s.FirewallID) {
				firewallName = fw.Metadata.Name
			}
		}

		// convert the pools
		pools := []*opencpspec.KubernetesClusterPool{}
		for _, pool := range k8s.Pools {
			pools = append(pools, &opencpspec.KubernetesClusterPool{
				Id:    pool.ID,
				Size:  pool.Size,
				Count: int32(pool.Count),
			})
		}

		kubernetesCluster = append(kubernetesCluster, &opencpspec.KubernetesCluster{
			Metadata: &metav1.ObjectMeta{
				Name:              k8s.Name,
				Namespace:         networkName,
				UID:               types.UID(k8s.ID),
				CreationTimestamp: metav1.Time{Time: k8s.CreatedAt},
			},
			Spec: &opencpspec.KubernetesClusterSpec{
				Pools:       pools,
				Version:     k8s.Version,
				Firewall:    firewallName,
				Cniplugin:   k8s.CNIPlugin,
				Clustertype: k8s.ClusterType,
				Kubeconfig:  k8s.KubeConfig,
			},
			Status: &opencpspec.KubernetesClusterStatus{
				State:    k8s.Status,
				Endpoint: k8s.APIEndPoint,
				Publicip: k8s.MasterIP,
			},
		})
	}

	// check if the filter is empty
	if option != nil {
		// check if namespace is set
		if option.Namespace != nil && *option.Namespace != "" {
			// filter the virtual machines by namespace
			filteredk8s := []*opencpspec.KubernetesCluster{}
			for _, k8s := range kubernetesCluster {
				if k8s.Metadata.Namespace == *option.Namespace {
					filteredk8s = append(filteredk8s, k8s)
				}
			}

			return &opencpspec.KubernetesClusterList{
				Items: filteredk8s,
			}, nil
		}
	}

	return &opencpspec.KubernetesClusterList{
		Items: kubernetesCluster,
	}, nil
}

func (s *Server) DeleteKubernetesCluster(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.KubernetesCluster, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Get the kubernetes cluster
	k8s, err := s.GetKubernetesCluster(ctx, option)
	if err != nil {
		return nil, err
	}

	// Delete the kubernetes cluster
	if k8s != nil {
		_, err = client.DeleteKubernetesCluster(string(k8s.Metadata.UID))
		if err != nil {
			return nil, err
		}
	}

	return k8s, nil
}
