package services

import (
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/store/cache"
)

func NewLister(store core.KVStore, selector *core.Service, handle *core.ResourceEventHandle) (core.ResourceLister, error) {
	prefix := getKey(selector)
	return cache.New(store, prefix, new(core.Service), handle)
}
