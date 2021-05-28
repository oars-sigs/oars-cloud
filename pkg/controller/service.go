package controller

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
	resStore "github.com/oars-sigs/oars-cloud/pkg/store/resources"

	log "github.com/sirupsen/logrus"
)

type serviceController struct {
	kv        core.KVStore
	trigger   chan struct{}
	svcLister core.ResourceLister
	edpLister core.ResourceLister
}

func newSvc(kv core.KVStore) *serviceController {
	trigger := make(chan struct{}, 1)
	return &serviceController{kv: kv, trigger: trigger}
}

func (c *serviceController) run() error {
	svcInterceptor := func(put bool, r, prer core.Resource) (core.Resource, bool, error) {
		res := false
		if r != nil {
			if r.(*core.Service).Kind == core.StaticServiceKind {
				res = true
			}
		}
		if prer != nil {
			if prer.(*core.Service).Kind == core.StaticServiceKind {
				res = true
			}
		}
		return nil, res, nil
	}
	svcLister, err := resStore.NewLister(c.kv, new(core.Service), &core.ResourceEventHandle{
		Interceptor: svcInterceptor,
		Trigger:     c.trigger,
	})
	if err != nil {
		return err
	}
	c.svcLister = svcLister

	edpInterceptor := func(put bool, r, prer core.Resource) (core.Resource, bool, error) {
		res := false
		if r != nil {
			if r.(*core.Endpoint).Kind == core.StaticServiceKind {
				res = true
			}
		}
		if prer != nil {
			if prer.(*core.Endpoint).Kind == core.StaticServiceKind {
				res = true
			}
		}
		return nil, res, nil
	}
	edpLister, err := resStore.NewLister(c.kv, new(core.Endpoint), &core.ResourceEventHandle{Interceptor: edpInterceptor})
	if err != nil {
		return err
	}
	c.edpLister = edpLister
	go c.handle()
	c.scheduler()
	return nil
}

func (c *serviceController) scheduler() {
	select {
	case c.trigger <- struct{}{}:
	default:
	}
}

type staticEndpoint struct {
	name string
	ip   string
}

func (c *serviceController) handle() {
	for {
		select {
		case <-c.trigger:
			svcRes, sok := c.svcLister.List()
			if !sok {
				continue
			}
			edpRes, eok := c.edpLister.List()
			if !eok {
				continue
			}
			addEdps := make([]*core.Endpoint, 0)
			delEdps := make([]*core.Endpoint, 0)
			//svc load
			for _, res := range svcRes {
				svc := res.(*core.Service)
				edpns := make([]*staticEndpoint, 0)
				edps := make([]*core.Endpoint, 0)

				for _, svcE := range svc.Static.Endpoints {
					se := &staticEndpoint{
						name: svcE.Name,
						ip:   svcE.IP,
					}
					edpns = append(edpns, se)
				}
				for _, res := range edpRes {
					edp := res.(*core.Endpoint)
					if edp.Service == svc.Name && edp.Namespace == svc.Namespace {
						edps = append(edps, edp)
					}
				}

				for _, edpn := range edpns {
					exist := false
					for _, edp := range edps {
						if edp.Name == edpn.name {
							exist = true
						}
					}
					if !exist {
						addEdps = append(addEdps, &core.Endpoint{
							ResourceMeta: &core.ResourceMeta{
								Name:      edpn.name,
								Namespace: svc.Namespace,
							},
							Service: svc.Name,
							Kind:    core.StaticServiceKind,
							Status: &core.EndpointStatus{
								IP:    edpn.ip,
								State: "running",
							},
						})
					}
				}

				for _, edp := range edps {
					exist := false
					for _, edpn := range edpns {
						if edp.Name == edpn.name {
							exist = true
						}
					}
					if !exist {
						delEdps = append(delEdps, edp)
					}
				}

			}

			//edp load
			for _, res := range edpRes {
				edp := res.(*core.Endpoint)
				exist := false
				for _, res := range svcRes {
					svc := res.(*core.Service)
					if edp.Service == svc.Name && edp.Namespace == svc.Namespace {
						exist = true
					}
				}
				if !exist {
					delEdps = append(delEdps, edp)
				}
			}
			//update edps
			edpStore := resStore.NewStore(c.kv, new(core.Endpoint))
			for _, edp := range addEdps {
				_, err := edpStore.Put(context.TODO(), edp, &core.PutOptions{})
				if err != nil {
					log.Error(err)
				}
			}
			for _, edp := range delEdps {
				err := edpStore.Delete(context.TODO(), edp, &core.DeleteOptions{})
				if err != nil {
					log.Error(err)
				}
			}
		}
	}
}
