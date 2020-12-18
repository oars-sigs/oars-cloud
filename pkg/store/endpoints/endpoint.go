package endpoints

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

type edpstore struct {
	store core.KVStore
}

//New endpoint store
func New(store core.KVStore) core.EndpointStore {
	return &edpstore{store}
}

func (s *edpstore) List(ctx context.Context, arg *core.Endpoint, opts *core.ListOptions) ([]*core.Endpoint, error) {
	endpoints := make([]*core.Endpoint, 0)
	key := getPrefixKey(arg)
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return endpoints, err
	}
	for _, kv := range kvs {
		endpoint := new(core.Endpoint)
		endpoint.Parse(kv.Value)
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func (s *edpstore) Get(ctx context.Context, arg *core.Endpoint, opts *core.GetOptions) (*core.Endpoint, error) {
	key := getKey(arg)
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return nil, err
	}
	if len(kvs) == 0 {
		return nil, e.ErrResourceNotFound
	}
	endpoint := new(core.Endpoint)
	endpoint.Parse(kvs[0].Value)
	return endpoint, nil
}

func (s *edpstore) Put(ctx context.Context, arg *core.Endpoint, opts *core.PutOptions) (*core.Endpoint, error) {
	endpoint, err := s.Get(ctx, arg, &core.GetOptions{})
	if err != nil {
		if err != e.ErrResourceNotFound {
			return arg, err
		}
		endpoint = new(core.Endpoint)
		endpoint.Created = time.Now().Unix()
	}
	arg.Updated = time.Now().Unix()
	arg.Created = endpoint.Created
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

func (s *edpstore) Delete(ctx context.Context, arg *core.Endpoint, opts *core.DeleteOptions) error {
	key := getKey(arg)
	err := s.store.Delete(ctx, key, core.KVOption{WithPrefix: true})
	return err
}
