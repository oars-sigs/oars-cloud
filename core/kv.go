package core

import (
	"context"
)

// KVOption action option
type KVOption struct {
	DisableFirst bool
	WithPrefix   bool
	WithPrevKV   bool
	WithRev      int64
}

// KV kv struct
type KV struct {
	Key   string
	Value string
}

// WatchChan watch chan
type WatchChan struct {
	Put    bool
	PrevKV KV
	KV     KV
}

//KVRegister 注册
type KVRegister interface {
	Close() error
}

//KVStore kv 存储
type KVStore interface {
	Put(ctx context.Context, kv KV) error
	Get(ctx context.Context, key string, op KVOption) ([]KV, error)
	GetWithRev(ctx context.Context, key string, op KVOption) ([]KV, int64, error)
	Delete(ctx context.Context, key string, op KVOption) error
	Watch(ctx context.Context, key string, updateCh chan WatchChan, errCh chan error, op KVOption)
	Register(ctx context.Context, kv KV, lease int64) (KVRegister, error)
	LeaderController(token string) LeaderController
}

//LeaderWorker leader worker
type LeaderWorker interface {
	Start()
	Stop()
}

//LeaderController
type LeaderController interface {
	Register(worker LeaderWorker)
	Close()
}
