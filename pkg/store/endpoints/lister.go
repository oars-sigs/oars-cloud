package endpoints

import (
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/store/cache"
)

func NewLister(store core.KVStore, selector *core.Endpoint, handle *core.ResourceEventHandle) (core.ResourceLister, error) {
	prefix := getPrefixKey(selector)
	return cache.New(store, prefix, new(core.Endpoint), handle)
}
