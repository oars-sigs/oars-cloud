package etcd

import (
	"context"
	"strings"

	"github.com/oars-sigs/oars-cloud/core"
	clientv3 "go.etcd.io/etcd/client/v3"
)

//Put puts a key-value pair into etcd.
func (s *Storage) Put(ctx context.Context, kv core.KV) error {
	key := s.keyPrefix + "/" + kv.Key
	ctx, cancel := s.newEtcdTimeoutContext(ctx)
	defer cancel()

	_, err := s.client.Put(ctx, key, kv.Value)
	return err
}

//Get get keys.
func (s *Storage) Get(ctx context.Context, key string, op core.KVOption) ([]core.KV, error) {
	key = s.keyPrefix + "/" + key
	ctx, cancel := s.newEtcdTimeoutContext(ctx)
	defer cancel()
	opts := []clientv3.OpOption{}
	if op.WithPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}

	gresp, err := s.client.Get(ctx, key, opts...)
	if err != nil {
		return nil, err
	}
	res := make([]core.KV, 0)
	for _, v := range gresp.Kvs {
		kv := core.KV{
			Key:   strings.TrimPrefix(string(v.Key), s.keyPrefix+"/"),
			Value: string(v.Value),
		}
		res = append(res, kv)
	}
	return res, nil
}

//GetWithRev retrieves keys.
func (s *Storage) GetWithRev(ctx context.Context, key string, op core.KVOption) ([]core.KV, int64, error) {
	key = s.keyPrefix + "/" + key
	ctx, cancel := s.newEtcdTimeoutContext(ctx)
	defer cancel()
	opts := []clientv3.OpOption{}
	if op.WithPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}

	gresp, err := s.client.Get(ctx, key, opts...)
	if err != nil {
		return nil, 0, err
	}
	res := make([]core.KV, 0)
	for _, v := range gresp.Kvs {
		kv := core.KV{
			Key:   strings.TrimPrefix(string(v.Key), s.keyPrefix+"/"),
			Value: string(v.Value),
		}
		res = append(res, kv)
	}
	return res, gresp.Header.Revision, nil
}

// Delete delete key
func (s *Storage) Delete(ctx context.Context, key string, op core.KVOption) error {
	key = s.keyPrefix + "/" + key
	ctx, cancel := s.newEtcdTimeoutContext(ctx)
	defer cancel()
	opts := []clientv3.OpOption{}
	if op.WithPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}
	_, err := s.client.Delete(ctx, key, opts...)
	return err
}
