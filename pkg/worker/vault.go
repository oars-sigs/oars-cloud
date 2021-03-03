package worker

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

// VaultClient vault client
type VaultClient struct {
	client *api.Client
}

// newVault new vault client
func newVault(addr, token string) (*VaultClient, error) {
	client, err := api.NewClient(&api.Config{
		Address: addr,
	})
	if err != nil {
		return nil, err
	}
	client.SetToken(token)
	return &VaultClient{client}, nil
}

func (v *VaultClient) Get(path, key string) (string, error) {
	s, err := v.client.Logical().Read(path)
	if err != nil {
		return "", err
	}
	if v, ok := s.Data[key]; ok {
		return fmt.Sprint(v), nil
	}
	return "", nil
}
