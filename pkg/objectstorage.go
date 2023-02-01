package pkg

import (
	"context"
	"fmt"

	"github.com/civo/civogo"
	"k8s.io/apimachinery/pkg/types"

	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) CreateObjectStorage(ctx context.Context, in *opencpspec.ObjectStorage) (*opencpspec.ObjectStorage, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Create Object Storage config
	objectStorageConfig := &civogo.CreateObjectStoreRequest{
		Name:        in.Metadata.Name,
		MaxSizeGB:   int64(in.Spec.Size),
		AccessKeyID: in.Spec.StorageCredential,
		Region:      client.Region,
	}

	// Create object storage
	objStorage, err := client.NewObjectStore(objectStorageConfig)
	if err != nil {
		return nil, err
	}

	// Get the object storage
	objectStorage, err := s.GetObjectStorage(ctx, &opencpspec.FilterOptions{Name: &objStorage.Name})
	if err != nil {
		return nil, err
	}

	return objectStorage, nil
}

func (s *Server) DeleteObjectStorage(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.ObjectStorage, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Get object storage
	objectStorage, err := s.GetObjectStorage(ctx, option)
	if err != nil {
		return nil, err
	}

	if objectStorage != nil {
		// Delete object storage
		_, err = client.DeleteObjectStore(string(objectStorage.Metadata.UID))
		if err != nil {
			return nil, err
		}
	}
	return objectStorage, nil
}

func (s *Server) GetObjectStorage(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.ObjectStorage, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Get object storage
	objectStorage, err := client.FindObjectStore(*option.Name)
	if err != nil {
		return nil, err
	}

	// Convert to opencp spec object storage
	return &opencpspec.ObjectStorage{
		Metadata: &metav1.ObjectMeta{
			Name: objectStorage.Name,
			UID:  types.UID(objectStorage.ID),
		},
		Spec: &opencpspec.ObjectStorageSpec{
			Size:              int32(objectStorage.MaxSize),
			StorageCredential: objectStorage.OwnerInfo.Name,
		},
		Status: &opencpspec.ObjectStorageStatus{
			State:             objectStorage.Status,
			Endpoint:          fmt.Sprintf("https://%s/%s", objectStorage.BucketURL, objectStorage.Name),
			StorageCredential: objectStorage.OwnerInfo.Name,
		},
	}, nil
}

func (s *Server) ListObjectStorage(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.ObjectStorageList, error) {
	client := ctx.Value("client").(*civogo.Client)

	// List all object storage
	objectStorages, err := client.ListObjectStores()
	if err != nil {
		return nil, err
	}

	// Convert to opencp spec object storage
	objectStorageList := &opencpspec.ObjectStorageList{}
	for _, objectStorage := range objectStorages.Items {
		objectStorageList.Items = append(objectStorageList.Items, &opencpspec.ObjectStorage{
			Metadata: &metav1.ObjectMeta{
				Name: objectStorage.Name,
				UID:  types.UID(objectStorage.ID),
			},
			Spec: &opencpspec.ObjectStorageSpec{
				Size:              int32(objectStorage.MaxSize),
				StorageCredential: objectStorage.OwnerInfo.Name,
			},
			Status: &opencpspec.ObjectStorageStatus{
				State:             objectStorage.Status,
				Endpoint:          fmt.Sprintf("https://%s/%s", objectStorage.BucketURL, objectStorage.Name),
				StorageCredential: objectStorage.OwnerInfo.Name,
			},
		})
	}

	return objectStorageList, nil
}
