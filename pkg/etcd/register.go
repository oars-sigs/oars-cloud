package etcd

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
	"go.etcd.io/etcd/clientv3"
)

type register struct {
	leaseID clientv3.LeaseID
	client  *clientv3.Client
}

//Register 注册服务
func (s *Storage) Register(ctx context.Context, kv core.KV, lease int64) (core.KVRegister, error) {
	key := s.keyPrefix + "/" + kv.Key
	resp, err := s.client.Grant(context.Background(), lease)
	if err != nil {
		return nil, err
	}
	//注册服务并绑定租约
	_, err = s.client.Put(context.Background(), key, kv.Value, clientv3.WithLease(resp.ID))
	if err != nil {
		return nil, err
	}
	//设置续租 定期发送需求请求
	ch, err := s.client.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			<-ch
		}
	}()
	ser := &register{
		client:  s.client,
		leaseID: resp.ID,
	}

	return ser, nil
}

// Close 注销服务
func (s *register) Close() error {
	//撤销租约
	if _, err := s.client.Revoke(context.Background(), s.leaseID); err != nil {
		return err
	}
	return nil
}
