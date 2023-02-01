package pkg

import (
	"context"

	"github.com/civo/civogo"

	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (s *Server) ListDatabase(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.DatabaseList, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// List all databases
	dbList, err := client.ListDatabases()
	if err != nil {
		return nil, err
	}

	// Get all the networks
	network, err := s.ListNamespace(ctx, option)
	if err != nil {
		return nil, err
	}

	// Get all firewall from Civo
	firewall, err := s.ListFirewall(ctx, option)
    if err!= nil {
        return nil, err
    }

    // Get all the namespaces

	// convert the database to the opencp format
	var allDatabase []*opencpspec.Database
	for _, db := range dbList.Items {

		// Get the right network
		var networkName string
		for _, net := range network.Items {
			if net.Metadata.UID == types.UID(db.NetworkID) {
				networkName = net.Metadata.Name
			}
		}

		// Get the right firewall
        var firewallName string
		for _, fw := range firewall.Items {
			if fw.Metadata.UID == types.UID(db.FirewallID) {
                firewallName = fw.Metadata.Name
            }
        }

		allDatabase = append(allDatabase, &opencpspec.Database{
			Metadata: &metav1.ObjectMeta{
				Name:      db.Name,
				UID:       types.UID(db.ID),
				Namespace: networkName,
			},
			Spec: &opencpspec.DatabaseSpec{
				Nodes:         int32(db.Nodes),
				Size:          db.Size,
				Engine:        db.Software,
				EngineVersion: db.SoftwareVersion,
				Firewall:      firewallName,
			},
			Status: &opencpspec.DatabaseStatus{
				Username: db.Username,
				Password: db.Password,
				Port:     int32(db.Port),
				PublicIP: db.PublicIPv4,
				State:    db.Status,
			},
		})
	}

	// check if the filter is empty
	if option != nil {
		// check if namespace is set
		if option.Namespace != nil && *option.Namespace != "" {
			// filter the virtual machines by namespace
			filteredDB := []*opencpspec.Database{}
			for _, db := range allDatabase {
				if db.Metadata.Namespace == *option.Namespace {
					filteredDB = append(filteredDB, db)
				}
			}

			return &opencpspec.DatabaseList{
				Items: filteredDB,
			}, nil
		}
	}

	return &opencpspec.DatabaseList{
		Items: allDatabase,
	}, nil
}

func (s *Server) GetDatabase(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.Database, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Find the database
	db, err := client.FindDatabase(*option.Name)
	if err != nil {
		return nil, err
	}

	// find the network and filter the database
	var networkName string
	if option.Namespace != nil && *option.Namespace != "" {
		network, err := s.GetNamespace(ctx, &opencpspec.FilterOptions{Name: option.Namespace})
		if err != nil {
			return nil, err
		}

		// filter the firewall by network
		if db.NetworkID != string(network.Metadata.UID) {
			return nil, nil
		}

		// set the network name
		networkName = network.Metadata.Name
	}

	// Get the firewall
	firewall, err := s.GetFirewall(ctx, &opencpspec.FilterOptions{Name: &db.FirewallID})
    if err!= nil {
        return nil, err
    }

	// Return the db in the opencp format
	return &opencpspec.Database{
		Metadata: &metav1.ObjectMeta{
			Name:      db.Name,
			UID:       types.UID(db.ID),
			Namespace: networkName,
		},
		Spec: &opencpspec.DatabaseSpec{
			Nodes:         int32(db.Nodes),
			Size:          db.Size,
			Engine:        db.Software,
			EngineVersion: db.SoftwareVersion,
			Firewall:      firewall.Metadata.Name,
		},
		Status: &opencpspec.DatabaseStatus{
			Username: db.Username,
			Password: db.Password,
			Port:     int32(db.Port),
			PublicIP: db.PublicIPv4,
			State:    db.Status,
		},
	}, nil
}

func (s *Server) CreateDatabase(ctx context.Context, in *opencpspec.Database) (*opencpspec.Database, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// get the network
	network, err := s.GetNamespace(ctx, &opencpspec.FilterOptions{Name: &in.Metadata.Namespace})
	if err != nil {
		return nil, err
	}

	// Create a Database config from the civo lib
	dbConfig := &civogo.CreateDatabaseRequest{
		Name:       in.Metadata.Name,
		Size:       in.Spec.Size,
		NetworkID:  string(network.Metadata.UID),
		Nodes:      int(in.Spec.Nodes),
		Region:     client.Region,
	}

	// Check if the incoming database have firewall
	if in.Spec.Firewall!= "" {
        // Get the firewall object
		firewall, err := s.GetFirewall(ctx, &opencpspec.FilterOptions{Name: &in.Spec.Firewall})
		if err!= nil {
			return nil, err
		}
		dbConfig.FirewallID = string(firewall.Metadata.UID)
    }

	// Create the database using the civo client
	_, err = client.NewDatabase(dbConfig)
	if err != nil {
		return nil, err
	}

	// Return the database in the opencp format
	return s.GetDatabase(ctx, &opencpspec.FilterOptions{Name: &in.Metadata.Name, Namespace: &network.Metadata.Name})
}

func (s *Server) DeleteDatabase(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.Database, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Find the database using the opencp
	db, err := s.GetDatabase(ctx, option)
    if err!= nil {
        return nil, err
    }

    // Delete the database
    _, err = client.DeleteDatabase(string(db.Metadata.UID))
    if err!= nil {
		return nil, err
    }

	// Return the database in the opencp format
    return db, nil
}
