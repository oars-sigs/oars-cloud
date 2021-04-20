package controller

import (
	"sync"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/controller/ingress"
	"github.com/oars-sigs/oars-cloud/pkg/controller/ingress/envoy"
	"github.com/oars-sigs/oars-cloud/pkg/controller/ingress/nginx"
	"github.com/oars-sigs/oars-cloud/pkg/controller/ingress/traefik"
	resStore "github.com/oars-sigs/oars-cloud/pkg/store/resources"
	"github.com/oars-sigs/oars-cloud/pkg/utils/strvars"
)

type ingressController struct {
	store          core.KVStore
	cfg            *core.Config
	trigger        chan struct{}
	listenerLister core.ResourceLister
	routeLister    core.ResourceLister
	certLister     core.ResourceLister
	traefikHandle  core.IngressControllerHandle
	envoyHandle    core.IngressControllerHandle
	nginxHandle    core.IngressControllerHandle
}

func newIngress(store core.KVStore, cfg *core.Config) *ingressController {
	trigger := make(chan struct{}, 1)
	return &ingressController{store: store, cfg: cfg, trigger: trigger}
}

func (c *ingressController) run(stopCh <-chan struct{}) error {
	handle := &core.ResourceEventHandle{
		Trigger: c.trigger,
	}
	listenerLister, err := resStore.NewLister(c.store, new(core.IngressListener), handle)
	if err != nil {
		return err
	}
	c.listenerLister = listenerLister
	routeLister, err := resStore.NewLister(c.store, new(core.IngressRoute), handle)
	if err != nil {
		return err
	}
	c.routeLister = routeLister
	certLister, err := resStore.NewLister(c.store, &core.Certificate{}, handle)
	if err != nil {
		return err
	}
	c.certLister = certLister
	if strvars.ArrayContains(c.cfg.Ingress.Drives, core.IngressTraefikDrive) {
		c.traefikHandle = traefik.New(listenerLister, routeLister, certLister)
	}
	if strvars.ArrayContains(c.cfg.Ingress.Drives, core.IngressNginxDrive) {
		c.nginxHandle = nginx.New(listenerLister, routeLister, certLister)
	}
	c.envoyHandle = envoy.New(listenerLister, routeLister, certLister, c.cfg.Ingress.XDSPort)
	go ingress.HTTPServer(c.cfg.Ingress.HTTPPort)
	go c.update(stopCh)
	c.scheduler()
	return nil
}

func (c *ingressController) scheduler() {
	select {
	case c.trigger <- struct{}{}:
	default:
	}
}

func (c *ingressController) update(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case <-c.trigger:
			wg := new(sync.WaitGroup)
			if c.traefikHandle != nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					c.traefikHandle.UpdateHandle()
				}()
			}
			if c.envoyHandle != nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					c.envoyHandle.UpdateHandle()
				}()
			}
			if c.nginxHandle != nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					c.nginxHandle.UpdateHandle()
				}()
			}
			wg.Wait()
		}
	}
}
