package pkg

import (
	"context"

	"github.com/civo/civogo"
	"k8s.io/apimachinery/pkg/types"

	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) ListDomains(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.DomainList, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Get all the domains again and return them
	allDomains, err := client.ListDNSDomains()
	if err != nil {
		return nil, err
	}

	// Convert the domains to the opencp format
	var domains []*opencpspec.Domain
	for _, domain := range allDomains {
		// Get the records
		records, err := client.ListDNSRecords(domain.ID)
		if err != nil {
			return nil, err
		}

		// Convert the records
		var domainRecords []*opencpspec.Records
		for _, record := range records {
			priority := int32(record.Priority)
			ttl := int32(record.TTL)

			domainRecords = append(domainRecords, &opencpspec.Records{
				Name:     record.Name,
				Type:     string(record.Type),
				Value:    record.Value,
				TTL:      &ttl,
				Priority: &priority,
			})
		}

		domains = append(domains, &opencpspec.Domain{
			Metadata: &metav1.ObjectMeta{
				Name: domain.Name,
				UID:  types.UID(domain.ID),
			},
			Spec: &opencpspec.DomainSpec{
				Records: domainRecords,
			},
			Status: &opencpspec.DomainStatus{
				State: "Active",
			},
		})
	}

	return &opencpspec.DomainList{
		Items: domains,
	}, nil
}

func (s *Server) CreateDomain(ctx context.Context, in *opencpspec.Domain) (*opencpspec.Domain, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Create the Domain
	domainResult, err := client.CreateDNSDomain(in.Metadata.Name)
	if err != nil {
		return nil, err
	}

	// Create the records
	if len(in.Spec.Records) > 0 {
		for _, record := range in.Spec.Records {
			recordConfig := &civogo.DNSRecordConfig{
				Type:  civogo.DNSRecordType(record.Type),
				Name:  record.Name,
				Value: record.Value,
			}

			if record.TTL != nil {
				recordConfig.TTL = int(*record.TTL)
			}

			if record.Priority != nil {
				recordConfig.Priority = int(*record.Priority)
			}

			_, err := client.CreateDNSRecord(domainResult.ID, recordConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	// Return the domain
	return &opencpspec.Domain{
		Metadata: &metav1.ObjectMeta{
			Name: domainResult.Name,
			UID:  types.UID(domainResult.ID),
		},
		Spec: &opencpspec.DomainSpec{
			Records: in.Spec.Records,
		},
		Status: &opencpspec.DomainStatus{
			State: "Active",
		},
	}, nil
}

func (s *Server) GetDomain(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.Domain, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Get the domain
	domainResult, err := client.FindDNSDomain(*option.Name)
	if err != nil {
		return nil, err
	}

	// Get the records
	records, err := client.ListDNSRecords(domainResult.ID)
	if err != nil {
		return nil, err
	}

	// Convert the records to the opencp format
	var recordsList []*opencpspec.Records
	for _, record := range records {
		priority := int32(record.Priority)
		ttl := int32(record.TTL)

		recordsList = append(recordsList, &opencpspec.Records{
			Name:     record.Name,
			Type:     string(record.Type),
			Value:    record.Value,
			Priority: &priority,
			TTL:      &ttl,
		})
	}

	// Return the domain
	return &opencpspec.Domain{
		Metadata: &metav1.ObjectMeta{
			Name: domainResult.Name,
			UID:  types.UID(domainResult.ID),
			CreationTimestamp: metav1.Time{
				Time: metav1.Now().Time,
			},
		},
		Spec: &opencpspec.DomainSpec{
			Records: recordsList,
		},
		Status: &opencpspec.DomainStatus{
			State: "Active",
		},
	}, nil
}

func (s *Server) DeleteDomain(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.Domain, error) {
	// Civo client from the ctx
	client := ctx.Value("client").(*civogo.Client)

	// Get the domain
	domainResult, err := client.FindDNSDomain(*option.Name)
	if err != nil {
		return nil, err
	}

	// Delete the domain
	_, err = client.DeleteDNSDomain(domainResult)
	if err != nil {
		return nil, err
	}

	// Return the domain
	return &opencpspec.Domain{
		Metadata: &metav1.ObjectMeta{
			Name: domainResult.Name,
			UID:  types.UID(domainResult.ID),
			CreationTimestamp: metav1.Time{
				Time: metav1.Now().Time,
			},
		},
		Spec: &opencpspec.DomainSpec{
			Records: nil,
		},
		Status: &opencpspec.DomainStatus{
			State: "Deleted",
		},
	}, nil
}

// UpdateDomain(Domain) returns (Domain) {}
// DeleteDomain(ctx context.Context, *opencpspec.FilterOptions) (*opencpspec.Domain, error)
//
