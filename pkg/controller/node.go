package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	log "github.com/sirupsen/logrus"
)

type nodeController struct {
	store core.KVStore
	cache sync.Map
}

func (c *nodeController) runNodec(stopCh <-chan struct{}) error {
	checkStopCh := make(chan struct{})
	go c.healthCheck(checkStopCh)
	updateCh := make(chan core.WatchChan)
	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	//TODO: 封装成存储库
	go c.store.Watch(ctx, "services/endpoint/namespaces/system/node/", updateCh, errCh, core.KVOption{WithPrevKV: true, WithPrefix: true})
	for {
		select {
		case res := <-updateCh:
			kv := res.KV
			if !res.Put {
				kv = res.PrevKV
			}
			endpoint := new(core.Endpoint)
			err := endpoint.Parse(kv.Value)
			if err != nil {
				log.Error(err)
				continue
			}
			if res.Put {
				c.cache.Store(endpoint.Hostname, endpoint)
				continue
			}
			c.cache.Delete(endpoint)
		case err := <-errCh:
			fmt.Println(err)
		case <-stopCh:
			cancel()
			checkStopCh <- struct{}{}
			return nil
		}
	}

}

func (c *nodeController) healthCheck(stopCh <-chan struct{}) {
	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-t.C:
			c.cache.Range(func(k, v interface{}) bool {
				endpoint := v.(*core.Endpoint)
				if time.Now().Unix()-endpoint.Updated > 60 && endpoint.State == "running" {
					endpoint.State = "error"
					endpoint.Status = "health check timeout"
					kv := core.KV{
						Key:   fmt.Sprintf("services/endpoint/namespaces/%s/%s/%s", endpoint.Namespace, endpoint.Service, endpoint.Hostname),
						Value: endpoint.String(),
					}
					err := c.store.Put(context.TODO(), kv)
					if err != nil {
						log.Error(err)
					}
				}
				return true
			})
		case <-stopCh:
			break
		}
	}
}
