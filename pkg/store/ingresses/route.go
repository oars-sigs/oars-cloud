package ingresses

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

type routeStore struct {
	store core.KVStore
}

//NewRoute ingress route store
func NewRoute(store core.KVStore) core.IngressRouteStore {
	return &routeStore{store}
}

func (s *routeStore) List(ctx context.Context, arg *core.IngressRoute, opts *core.ListOptions) ([]*core.IngressRoute, error) {
	nss := make([]*core.IngressRoute, 0)
	key := getRouteKey(arg)
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return nss, err
	}
	for _, kv := range kvs {
		ns := new(core.IngressRoute)
		ns.Parse(kv.Value)
		nss = append(nss, ns)
	}
	return nss, nil
}

func (s *routeStore) Get(ctx context.Context, arg *core.IngressRoute, opts *core.GetOptions) (*core.IngressRoute, error) {
	key := getRouteKey(arg)
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return nil, err
	}
	if len(kvs) == 0 {
		return nil, e.ErrResourceNotFound
	}
	ns := new(core.IngressRoute)
	ns.Parse(kvs[0].Value)
	return ns, nil
}

func (s *routeStore) Put(ctx context.Context, arg *core.IngressRoute, opts *core.PutOptions) (*core.IngressRoute, error) {
	ns, err := s.Get(ctx, arg, &core.GetOptions{})
	if err != nil {
		if err != e.ErrResourceNotFound {
			return arg, err
		}
		ns = new(core.IngressRoute)
		ns.Created = time.Now().Unix()
	}
	arg.Updated = time.Now().Unix()
	arg.Created = ns.Created
	v := core.KV{
		Key:   getRouteKey(arg),
		Value: arg.String(),
	}
	err = s.store.Put(ctx, v)
	if err != nil {
		return arg, err
	}
	return arg, nil
}

func (s *routeStore) Delete(ctx context.Context, arg *core.IngressRoute, opts *core.DeleteOptions) error {
	key := getRouteKey(arg)
	err := s.store.Delete(ctx, key, core.KVOption{WithPrefix: true})
	return err
}
