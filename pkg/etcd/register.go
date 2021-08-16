package etcd

import (
	"context"
	"strings"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type register struct {
	client *clientv3.Client
	stopCh chan struct{}
}

//Register 注册服务
func (s *Storage) Register(ctx context.Context, kv core.KV, lease int64) (core.KVRegister, error) {
	key := s.keyPrefix + "/" + kv.Key
	ser := &register{
		client: s.client,
		stopCh: make(chan struct{}),
	}
	go ser.start(key, kv.Value, lease)
	return ser, nil
}

func (s *register) start(key, value string, lease int64) {
	var curLeaseId clientv3.LeaseID = 0
	t := time.NewTicker(time.Second)
	for {
		select {
		case <-t.C:
			if curLeaseId == 0 {
				resp, err := s.client.Grant(context.Background(), lease)
				if err != nil {
					continue
				}
				curLeaseId = resp.ID
				//注册服务并绑定租约
				_, err = s.client.Put(context.Background(), key, value, clientv3.WithLease(resp.ID))
				if err != nil {
					continue
				}
			} else {
				//设置续租 定期发送需求请求
				_, err := s.client.KeepAliveOnce(context.Background(), curLeaseId)
				if err != nil && strings.Contains(err.Error(), "requested lease not found") {
					curLeaseId = 0
				}
			}
		case <-s.stopCh:
			t.Stop()
			s.client.Revoke(context.Background(), curLeaseId)
		}
	}
}

// Close 注销服务
func (s *register) Close() error {
	s.stopCh <- struct{}{}
	return nil
}
