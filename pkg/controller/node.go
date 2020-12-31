package controller

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	resStore "github.com/oars-sigs/oars-cloud/pkg/store/resources"
	log "github.com/sirupsen/logrus"
)

type nodeController struct {
	kv        core.KVStore
	store     core.ResourceStore
	lister    core.ResourceLister
	regLister core.ResourceLister
}

func newNodec(kv core.KVStore) *nodeController {
	return &nodeController{kv: kv}
}

func (c *nodeController) runNodec(stopCh chan struct{}) error {
	c.store = resStore.NewStore(c.kv, new(core.Endpoint))
	edp := &core.Endpoint{
		ResourceMeta: &core.ResourceMeta{
			Namespace: "system",
		},
		Service: "node",
	}
	lister, err := resStore.NewLister(c.kv, edp, &core.ResourceEventHandle{})
	if err != nil {
		return err
	}
	c.lister = lister
	edpReg := &core.Endpoint{
		ResourceMeta: &core.ResourceMeta{
			Namespace: "system",
			ObjectKind: &core.ResourceObjectKind{
				IsRegister: true,
			},
		},
		Service: "node",
	}
	regLister, err := resStore.NewLister(c.kv, edpReg, &core.ResourceEventHandle{})
	if err != nil {
		return err
	}
	c.regLister = regLister

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
			regResources, ok := c.regLister.List()
			if !ok {
				continue
			}

			for _, resource := range resources {
				endpoint := resource.(*core.Endpoint)
				isExist := false
				for _, regResource := range regResources {
					regEdp := regResource.(*core.Endpoint)
					if regEdp.Name == endpoint.Name {
						isExist = true
					}
				}

				if !isExist && endpoint.Status.State == "running" {
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
