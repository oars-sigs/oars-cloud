package resources

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
)

type register struct {
	kvreg core.KVRegister
}

//NewRegister ...
func NewRegister(store core.KVStore, resource core.Resource, lease int64) (core.ResourceRegister, error) {
	v := core.KV{
		Key:   registerPrefixKey + getKey(resource),
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
