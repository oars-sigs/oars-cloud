package resources

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

type store struct {
	kvstore core.KVStore
	cur     core.Resource
}

//NewStore resource store
func NewStore(kvstore core.KVStore, cur core.Resource) core.ResourceStore {
	return &store{kvstore, cur}
}

func (s *store) List(ctx context.Context, arg core.Resource, opts *core.ListOptions) ([]core.Resource, error) {
	ress := make([]core.Resource, 0)
	key := getPrefixKey(arg)
	kvs, err := s.kvstore.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return ress, err
	}
	for _, kv := range kvs {
		res := s.cur.New()
		res.Parse(kv.Value)
		ress = append(ress, res)
	}
	return ress, nil
}

func (s *store) Get(ctx context.Context, arg core.Resource, opts *core.GetOptions) (core.Resource, error) {
	key := getKey(arg)
	kvs, err := s.kvstore.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return nil, err
	}
	if len(kvs) == 0 {
		return nil, e.ErrResourceNotFound
	}
	res := s.cur.New()
	res.Parse(kvs[0].Value)
	return res, nil
}

func (s *store) Put(ctx context.Context, arg core.Resource, opts *core.PutOptions) (core.Resource, error) {
	res, err := s.Get(ctx, arg, &core.GetOptions{})
	if err != nil {
		if err != e.ErrResourceNotFound {
			return arg, err
		}
		res = s.cur.New()
		res.SetCreated(time.Now().Unix())
	}
	arg.SetUpdated(time.Now().Unix())
	arg.SetCreated(res.GetCreated())
	v := core.KV{
		Key:   getKey(arg),
		Value: arg.String(),
	}
	err = s.kvstore.Put(ctx, v)
	if err != nil {
		return arg, err
	}
	return arg, nil
}

func (s *store) Delete(ctx context.Context, arg core.Resource, opts *core.DeleteOptions) error {
	key := getKey(arg)
	err := s.kvstore.Delete(ctx, key, core.KVOption{WithPrefix: true})
	return err
}

func getKey(arg core.Resource) string {
	return arg.ResourceGroup() + "/" + arg.ResourceKind() + "/" + arg.ResourceKey()
}

func getPrefixKey(arg core.Resource) string {
	return arg.ResourceGroup() + "/" + arg.ResourceKind() + "/" + arg.ResourcePrefixKey()
}
