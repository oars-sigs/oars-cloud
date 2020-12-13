package services

import (
	"context"
	"fmt"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

type edpstore struct {
	store core.KVStore
}

//New endpoint store
func New(store core.KVStore) core.ServiceStore {
	return &edpstore{store}
}

func (s *edpstore) List(ctx context.Context, arg *core.Service, opts *core.ListOptions) ([]*core.Service, error) {
	endpoints := make([]*core.Service, 0)
	key := fmt.Sprintf("services/endpoint/namespaces/%s/%s", arg.Namespace, arg.Name)
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return endpoints, err
	}
	for _, kv := range kvs {
		endpoint := new(core.Service)
		endpoint.Parse(kv.Value)
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func (s *edpstore) Get(ctx context.Context, arg *core.Service, opts *core.GetOptions) (*core.Service, error) {
	key := fmt.Sprintf("services/svc/namespaces/%s/%s/%s", arg.Namespace, arg.Name)
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return nil, err
	}
	if len(kvs) == 0 {
		return nil, e.ErrResourceNotFound
	}
	endpoint := new(core.Service)
	endpoint.Parse(kvs[0].Value)
	return endpoint, nil
}

func (s *edpstore) PutSpec(ctx context.Context, arg *core.Service, opts *core.PutOptions) (*core.Service, error) {
	endpoint, err := s.Get(ctx, arg, &core.GetOptions{})
	if err != nil {
		if err != e.ErrResourceNotFound {
			return arg, err
		}
		endpoint = new(core.Service)
		endpoint.Created = time.Now().Unix()
	}
	arg.Updated = time.Now().Unix()
	arg.Created = endpoint.Updated
	v := core.KV{
		Key:   fmt.Sprintf("services/endpoint/namespaces/%s/%s", arg.Namespace, arg.Name),
		Value: arg.String(),
	}
	err = s.store.Put(ctx, v)
	if err != nil {
		return arg, err
	}
	return arg, nil
}
