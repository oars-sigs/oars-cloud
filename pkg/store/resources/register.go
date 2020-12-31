package resources

import (
	"context"
	"strings"

	"github.com/oars-sigs/oars-cloud/core"
)

type register struct {
	kvreg core.KVRegister
}

//NewRegister ...
func NewRegister(store core.KVStore, resource core.Resource, lease int64) (core.ResourceRegister, error) {
	key := getKey(resource)
	if !strings.HasPrefix(key, registerPrefixKey) {
		key = registerPrefixKey + getKey(resource)
	}
	v := core.KV{
		Key:   key,
		Value: resource.String(),
	}
	reg, err := store.Register(context.Background(), v, lease)
	if err != nil {
		return nil, err
	}
	return &register{
		kvreg: reg,
	}, nil
}

func (r *register) Close() error {
	return r.kvreg.Close()
}
