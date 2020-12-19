package resources

import (
	"context"
	"sync"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	log "github.com/sirupsen/logrus"
)

type client struct {
	store    core.KVStore
	data     map[string]core.Resource
	resource core.Resource
	mu       *sync.Mutex
	ready    bool
	handle   *core.ResourceEventHandle
}

//NewLister resource lister
func NewLister(store core.KVStore, resource core.Resource, handle *core.ResourceEventHandle) (core.ResourceLister, error) {
	c := &client{
		store:    store,
		resource: resource,
		handle:   handle,
		mu:       new(sync.Mutex),
	}
	stopCh := make(chan struct{})
	go func() {
		for {
			rev, err := c.fetch()
			if err != nil {
				log.Error(err)
				time.Sleep(time.Second * 5)
				continue
			}
			c.ready = true
			err = c.watch(rev, stopCh)
			if err == nil {
				break
			}
			time.Sleep(time.Second * 5)
		}
	}()
	return c, nil
}

func (c *client) List() ([]core.Resource, bool) {
	res := make([]core.Resource, 0)
	if !c.ready {
		return res, false
	}
	c.mu.Lock()
	for _, v := range c.data {
		res = append(res, v)
	}
	c.mu.Unlock()
	return res, true
}

func (c *client) fetch() (int64, error) {
	kvs, rev, err := c.store.GetWithRev(context.Background(), getPrefixKey(c.resource), core.KVOption{WithPrefix: true})
	if err != nil {
		return rev, err
	}
	ress := make(map[string]core.Resource, 0)
	for _, kv := range kvs {
		resource, ok, err := c.parseResource(true, kv, core.KV{})
		if err != nil {
			return 0, err
		}
		if !ok {
			continue
		}
		ress[resource.ResourceKey()] = resource
	}
	for _, oldRes := range c.data {
		if c.handle.Interceptor != nil {
			_, _, err := c.handle.Interceptor(false, nil, oldRes)
			if err != nil {
				return 0, err
			}
		}
	}
	c.mu.Lock()
	c.data = ress
	c.mu.Unlock()
	c.scheduler()
	return rev, nil
}

func (c *client) parseResource(put bool, kv, prekv core.KV) (core.Resource, bool, error) {
	var resource core.Resource
	if kv.Value != "" {
		resource = c.resource.New()
		err := resource.Parse(kv.Value)
		if err != nil {
			log.Error(err)
			//ignore parse error
			return resource, false, nil
		}
	}

	var preResource core.Resource
	if prekv.Value != "" {
		preResource = c.resource.New()
		err := preResource.Parse(prekv.Value)
		if err != nil {
			log.Error(err)
			//ignore parse error
			return resource, false, nil
		}
	}
	if c.handle.Interceptor != nil {
		res, ok, err := c.handle.Interceptor(put, resource, preResource)
		if err != nil {
			log.Error(err)
			return resource, false, err
		}
		if !ok {
			return resource, false, nil
		}
		if res != nil {
			resource = res
		}
	}
	if resource == nil {
		resource = preResource
	}
	return resource, true, nil
}

func (c *client) scheduler() {
	if c.handle.Trigger == nil {
		return
	}
	select {
	case c.handle.Trigger <- struct{}{}:
	default:
	}
}

func (c *client) watch(rev int64, stopCh chan struct{}) error {
	updateCh := make(chan core.WatchChan)
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	opt := core.KVOption{WithPrevKV: true, WithPrefix: true, DisableFirst: true, WithRev: rev}
	go c.store.Watch(ctx, getPrefixKey(c.resource), updateCh, errCh, opt)
	for {
		select {
		case res := <-updateCh:
			resource, ok, err := c.parseResource(res.Put, res.KV, res.PrevKV)
			if err != nil {
				continue
			}
			if !ok {
				continue
			}
			c.mu.Lock()
			if res.Put {
				c.data[resource.ResourceKey()] = resource

			} else {
				delete(c.data, resource.ResourceKey())
			}
			c.mu.Unlock()
			c.scheduler()

		case err := <-errCh:
			cancel()
			log.Error(err)
			return err
		case <-stopCh:
			cancel()
			return nil
		}
	}
}
