package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/oars-sigs/oars-cloud/core"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
)

// Storage stroage with etcd
type Storage struct {
	client     *clientv3.Client
	keyPrefix  string
	reqTimeout time.Duration
	cfg        *core.EtcdConfig
}

// New init a etcd client
func New(cfg *core.EtcdConfig, timeout time.Duration) (core.KVStore, error) {
	etcdCfg := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: timeout,
	}
	if cfg.TLS {
		tlsInfo := transport.TLSInfo{
			CertFile:      cfg.CertFile,
			KeyFile:       cfg.KeyFile,
			TrustedCAFile: cfg.TrustedCAFile,
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}
		etcdCfg.TLS = tlsConfig
	}
	client, err := clientv3.New(etcdCfg)
	if err != nil {
		return nil, err
	}

	return &Storage{
		client:     client,
		keyPrefix:  cfg.Prefix,
		reqTimeout: timeout,
		cfg:        cfg,
	}, nil
}

// etcdTimeoutContext etcd timeout context
type etcdTimeoutContext struct {
	context.Context

	etcdEndpoints []string
}

// Err err
func (c *etcdTimeoutContext) Err() error {
	err := c.Context.Err()
	if err == context.DeadlineExceeded {
		err = fmt.Errorf("%s: etcd(%v) lost",
			err, c.etcdEndpoints)
	}
	return err
}

// newEtcdTimeoutContext new etcd timeout context
func (s *Storage) newEtcdTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, s.reqTimeout)
	etcdCtx := &etcdTimeoutContext{}
	etcdCtx.Context = ctx
	etcdCtx.etcdEndpoints = s.client.Endpoints()
	return etcdCtx, cancel
}
