package namespaces

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

type nsstore struct {
	store core.KVStore
}

//New endpoint store
func New(store core.KVStore) core.NamespaceStore {
	return &nsstore{store}
}

func (s *nsstore) List(ctx context.Context, arg *core.Namespace, opts *core.ListOptions) ([]*core.Namespace, error) {
	nss := make([]*core.Namespace, 0)
	key := getKey(arg)
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return nss, err
	}
	for _, kv := range kvs {
		ns := new(core.Namespace)
		ns.Parse(kv.Value)
		nss = append(nss, ns)
	}
	return nss, nil
}

func (s *nsstore) Get(ctx context.Context, arg *core.Namespace, opts *core.GetOptions) (*core.Namespace, error) {
	key := getKey(arg)
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return nil, err
	}
	if len(kvs) == 0 {
		return nil, e.ErrResourceNotFound
	}
	ns := new(core.Namespace)
	ns.Parse(kvs[0].Value)
	return ns, nil
}

func (s *nsstore) Put(ctx context.Context, arg *core.Namespace, opts *core.PutOptions) (*core.Namespace, error) {
	ns, err := s.Get(ctx, arg, &core.GetOptions{})
	if err != nil {
		if err != e.ErrResourceNotFound {
			return arg, err
		}
		ns = new(core.Namespace)
		ns.Created = time.Now().Unix()
	}
	arg.Updated = time.Now().Unix()
	arg.Created = ns.Created
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

func (s *nsstore) Delete(ctx context.Context, arg *core.Namespace, opts *core.DeleteOptions) error {
	key := getKey(arg)
	err := s.store.Delete(ctx, key, core.KVOption{WithPrefix: true})
	return err
}
