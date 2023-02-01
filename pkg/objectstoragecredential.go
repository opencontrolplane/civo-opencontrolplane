package pkg

import (
	"context"
	"github.com/civo/civogo"
	"k8s.io/apimachinery/pkg/types"

	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) CreateObjectStorageCredential(ctx context.Context, in *opencpspec.ObjectStorageCredential) (*opencpspec.ObjectStorageCredential, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Create Object Storage config
	objectStorageCredentialConfig := &civogo.CreateObjectStoreCredentialRequest{
		Name:              in.Metadata.Name,
		Region:            client.Region,
	}

	// Check if the access key is provided
	if in.Spec.Accesskey != "" {
		objectStorageCredentialConfig.AccessKeyID = &in.Spec.Accesskey
	}

	// Check if the secret key is provided
	if in.Spec.Secretkey != "" {
		objectStorageCredentialConfig.SecretAccessKeyID = &in.Spec.Secretkey
	}

	// Create object storage
	objStorage, err := client.NewObjectStoreCredential(objectStorageCredentialConfig)
	if err != nil {
		return nil, err
	}

	// Get the object storage
	objectStorage, err := s.GetObjectStorageCredential(ctx, &opencpspec.FilterOptions{Name: &objStorage.Name})
	if err != nil {
		return nil, err
	}

	return objectStorage, nil
}

func (s *Server) DeleteObjectStorageCredential(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.ObjectStorageCredential, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Get object storage
	objectStorageCredential, err := s.GetObjectStorageCredential(ctx, option)
	if err != nil {
		return nil, err
	}

	if objectStorageCredential != nil {
		// Delete object storage
		_, err = client.DeleteObjectStore(string(objectStorageCredential.Metadata.UID))
		if err != nil {
			return nil, err
		}
	}
	return objectStorageCredential, nil
}

func (s *Server) GetObjectStorageCredential(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.ObjectStorageCredential, error) {
	client := ctx.Value("client").(*civogo.Client)

	// Get object storage
	objectStorageCredential, err := client.FindObjectStoreCredential(*option.Name)
	if err != nil {
		return nil, err
	}

	// Convert to opencp spec object storage
	return &opencpspec.ObjectStorageCredential{
		Metadata: &metav1.ObjectMeta{
			Name: objectStorageCredential.Name,
			UID:  types.UID(objectStorageCredential.ID),
		},
		Spec: &opencpspec.ObjectStorageCredentialSpec{
			Accesskey: objectStorageCredential.AccessKeyID,
			Secretkey: objectStorageCredential.SecretAccessKeyID,
		},
		Status: &opencpspec.ObjectStorageCredentialStatus{
			State: objectStorageCredential.Status,
		},
	}, nil
}

func (s *Server) ListObjectStorageCredential(ctx context.Context, option *opencpspec.FilterOptions) (*opencpspec.ObjectStorageCredentialList, error) {
	client := ctx.Value("client").(*civogo.Client)

	// List all object storage
	objectStoragesCredential, err := client.ListObjectStoreCredentials()
	if err != nil {
		return nil, err
	}

	// Convert to opencp spec object storage credential list
	objectStorageCredentialList := &opencpspec.ObjectStorageCredentialList{}
	for _, creadential := range objectStoragesCredential.Items {
		objectStorageCredentialList.Items = append(objectStorageCredentialList.Items, &opencpspec.ObjectStorageCredential{
			Metadata: &metav1.ObjectMeta{
				Name: creadential.Name,
				UID:  types.UID(creadential.ID),
			},
			Spec: &opencpspec.ObjectStorageCredentialSpec{
				Accesskey: creadential.AccessKeyID,
				Secretkey: creadential.SecretAccessKeyID,
			},
			Status: &opencpspec.ObjectStorageCredentialStatus{
				State: creadential.Status,
			},
		})
	}

	return objectStorageCredentialList, nil
}
