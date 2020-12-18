package services

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

type svcstore struct {
	store core.KVStore
}

//New endpoint store
func New(store core.KVStore) core.ServiceStore {
	return &svcstore{store}
}

func (s *svcstore) List(ctx context.Context, arg *core.Service, opts *core.ListOptions) ([]*core.Service, error) {
	endpoints := make([]*core.Service, 0)
	key := getPrefixKey(arg)
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

func (s *svcstore) Get(ctx context.Context, arg *core.Service, opts *core.GetOptions) (*core.Service, error) {
	key := getKey(arg)
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

func (s *svcstore) Put(ctx context.Context, arg *core.Service, opts *core.PutOptions) (*core.Service, error) {
	svc, err := s.Get(ctx, arg, &core.GetOptions{})
	if err != nil {
		if err != e.ErrResourceNotFound {
			return arg, err
		}
		svc = new(core.Service)
		svc.Created = time.Now().Unix()
	}
	arg.Updated = time.Now().Unix()
	arg.Created = svc.Created
	v := core.KV{
		Key:   getKey(arg),
		Value: arg.String(),
	}
	err = s.store.Put(ctx, v)
	if err != nil {
		return arg, err
	}
	return arg, nil
}

func (s *svcstore) Delete(ctx context.Context, arg *core.Service, opts *core.DeleteOptions) error {
	key := getKey(arg)
	err := s.store.Delete(ctx, key, core.KVOption{WithPrefix: true})
	return err
}
