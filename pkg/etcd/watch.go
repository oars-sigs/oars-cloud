package etcd

import (
	"context"
	"strings"

	"github.com/oars-sigs/oars-cloud/core"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Watch watches on a key or prefix
func (s *Storage) Watch(ctx context.Context, key string, updateCh chan core.WatchChan, errCh chan error, op core.KVOption) {
	key = s.keyPrefix + "/" + key
	opts := make([]clientv3.OpOption, 0)
	if op.WithPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}
	if op.WithPrevKV {
		opts = append(opts, clientv3.WithPrevKV())
	}
	gctx, cancel := s.newEtcdTimeoutContext(ctx)
	defer cancel()

	if !op.DisableFirst {
		gresp, err := s.client.Get(gctx, key, clientv3.WithPrefix())
		if err != nil {
			errCh <- err
		} else {
			for _, k := range gresp.Kvs {
				kv := core.KV{
					Key:   strings.TrimPrefix(string(k.Key), s.keyPrefix+"/"),
					Value: string(k.Value),
				}
				watch := core.WatchChan{
					Put:    true,
					PrevKV: kv,
					KV:     kv,
				}
				updateCh <- watch
			}
		}
		opts = append(opts, clientv3.WithRev(gresp.Header.Revision+1))
	} else {
		opts = append(opts, clientv3.WithRev(op.WithRev+1))
	}

	wch := s.client.Watch(context.Background(), key, opts...)
	for {
		select {
		case c := <-wch:
			if c.Err() != nil {
				errCh <- c.Err()
				continue
			}
			for _, e := range c.Events {
				isPut := false
				if e.Type == clientv3.EventTypePut {
					isPut = true
				}
				preV := ""
				if e.PrevKv != nil {
					preV = string(e.PrevKv.Value)
				}
				prev := core.KV{
					Key:   strings.TrimPrefix(string(e.Kv.Key), s.keyPrefix+"/"),
					Value: preV,
				}
				kv := core.KV{
					Key:   strings.TrimPrefix(string(e.Kv.Key), s.keyPrefix+"/"),
					Value: string(e.Kv.Value),
				}
				watch := core.WatchChan{
					Put:    isPut,
					PrevKV: prev,
					KV:     kv,
				}
				updateCh <- watch
			}
		case <-ctx.Done():
			return
		}
	}
}
