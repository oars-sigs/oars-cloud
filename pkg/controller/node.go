package controller

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	edpStore "github.com/oars-sigs/oars-cloud/pkg/store/endpoints"
	log "github.com/sirupsen/logrus"
)

type nodeController struct {
	kv     core.KVStore
	store  core.EndpointStore
	lister core.ResourceLister
}

func newNodec(kv core.KVStore) *nodeController {
	return &nodeController{kv: kv}
}

func (c *nodeController) runNodec(stopCh chan struct{}) error {
	c.store = edpStore.New(c.kv)
	lister, err := edpStore.NewLister(c.kv, &core.Endpoint{Namespace: "admin", Service: "node"}, &core.ResourceEventHandle{})
	if err != nil {
		return err
	}
	c.lister = lister
	checkStopCh := make(chan struct{})
	c.healthCheck(checkStopCh)
	return nil
}

func (c *nodeController) healthCheck(stopCh <-chan struct{}) {
	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-t.C:
			resources, ok := c.lister.List()
			if !ok {
				continue
			}
			for _, resource := range resources {
				endpoint := resource.(*core.Endpoint)
				if time.Now().Unix()-endpoint.Updated > 60 && endpoint.Status.State == "running" {
					endpoint.Status.State = "error"
					endpoint.Status.StateDetail = "health check timeout"
					_, err := c.store.Put(context.Background(), endpoint, &core.PutOptions{})
					if err != nil {
						log.Error(err)
					}
				}
			}
		case <-stopCh:
			break
		}
	}
}
