package ingresses

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

type listenerStore struct {
	store core.KVStore
}

//NewListener ingress listener store
func NewListener(store core.KVStore) core.IngressListenerStore {
	return &listenerStore{store}
}

func (s *listenerStore) List(ctx context.Context, arg *core.IngressListener, opts *core.ListOptions) ([]*core.IngressListener, error) {
	nss := make([]*core.IngressListener, 0)
	key := getListenerKey(arg)
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return nss, err
	}
	for _, kv := range kvs {
		ns := new(core.IngressListener)
		ns.Parse(kv.Value)
		nss = append(nss, ns)
	}
	return nss, nil
}

func (s *listenerStore) Get(ctx context.Context, arg *core.IngressListener, opts *core.GetOptions) (*core.IngressListener, error) {
	key := getListenerKey(arg)
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return nil, err
	}
	if len(kvs) == 0 {
		return nil, e.ErrResourceNotFound
	}
	ns := new(core.IngressListener)
	ns.Parse(kvs[0].Value)
	return ns, nil
}

func (s *listenerStore) Put(ctx context.Context, arg *core.IngressListener, opts *core.PutOptions) (*core.IngressListener, error) {
	ns, err := s.Get(ctx, arg, &core.GetOptions{})
	if err != nil {
		if err != e.ErrResourceNotFound {
			return arg, err
		}
		ns = new(core.IngressListener)
		ns.Created = time.Now().Unix()
	}
	arg.Updated = time.Now().Unix()
	arg.Created = ns.Created
	v := core.KV{
		Key:   getListenerKey(arg),
		Value: arg.String(),
	}
	err = s.store.Put(ctx, v)
	if err != nil {
		return arg, err
	}
	return arg, nil
}

func (s *listenerStore) Delete(ctx context.Context, arg *core.IngressListener, opts *core.DeleteOptions) error {
	key := getListenerKey(arg)
	err := s.store.Delete(ctx, key, core.KVOption{WithPrefix: true})
	return err
}
