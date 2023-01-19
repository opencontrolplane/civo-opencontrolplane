package main

import (
	"context"
	opencpspec "github.com/opencontrolplane/opencp-spec/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// TestSSHKeyList tests the ListSSHKey function
func TestSSHKeyList(t *testing.T) {
	ctx := context.Background()

	client, closer := FakeServer(ctx)
	defer closer()

	type expectation struct {
		out *opencpspec.SSHKeyList
		err error
	}

	searchValue := "opencp"

	tests := map[string]struct {
		in       *opencpspec.FilterOptions
		expected expectation
	}{
		"Must_Success": {
			in: &opencpspec.FilterOptions{
				Name: &searchValue,
			},
			expected: expectation{
				out: &opencpspec.SSHKeyList{
					Items: []*opencpspec.SSHKey{
						{
							Metadata: &metav1.ObjectMeta{
								Name: "opencp",
								UID:  "opencp",
							},
							Spec: &opencpspec.SSHKeySpec{
								Publickey: "test",
							},
							Status: &opencpspec.SSHKeyStatus{
								Fingerprint: "test",
								State:       "Active",
							},
						},
					},
				},
				err: nil,
			},
		},
	}

	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			out, err := client.ListSSHKey(ctx, tt.in)
			if err != nil {
				if tt.expected.err.Error() != err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.err, err)
				}
			} else {
				if len(tt.expected.out.Items) == 0 {
					t.Errorf("Out -> \nWant: %q\nGot : %q", tt.expected.out, out)
				}
			}

		})
	}
}
