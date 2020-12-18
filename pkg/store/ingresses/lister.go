package ingresses

import (
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/store/cache"
)

func NewListenerLister(store core.KVStore, selector *core.IngressListener, handle *core.ResourceEventHandle) (core.ResourceLister, error) {
	prefix := getListenerKey(selector)
	return cache.New(store, prefix, new(core.IngressListener), handle)
}

func NewRouteLister(store core.KVStore, selector *core.IngressRoute, handle *core.ResourceEventHandle) (core.ResourceLister, error) {
	prefix := getRoutePrefixKey(selector)
	return cache.New(store, prefix, new(core.IngressRoute), handle)
}
