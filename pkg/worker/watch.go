package worker

import (
	"context"
	"os"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	resStore "github.com/oars-sigs/oars-cloud/pkg/store/resources"
	"github.com/sirupsen/logrus"
)

func (d *daemon) run() {
	err := d.cacheService()
	if err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
	go d.cacheContainers()
	t := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-t.C:
			_, ok := d.svcLister.List()
			if !ok {
				logrus.Error("svc lister not ready")
				continue
			}
			if !d.ready {
				logrus.Error("container lister not ready")
				continue
			}
			d.syncDockerSvc()
		}
	}
}

func (d *daemon) cacheEndpoint() error {
	edpLister, err := resStore.NewLister(d.store, &core.Endpoint{}, &core.ResourceEventHandle{})
	if err != nil {
		return err
	}
	d.edpLister = edpLister
	interceptor := func(put bool, r, prer core.Resource) (core.Resource, bool, error) {
		res := false
		if r != nil {
			if r.(*core.Endpoint).Status.Node.Hostname == d.node.Hostname && r.(*core.Endpoint).Kind == "container" {
				res = true
			}
		}
		if prer != nil {
			if prer.(*core.Endpoint).Status.Node.Hostname == d.node.Hostname && prer.(*core.Endpoint).Kind == "container" {
				res = true
			}
		}
		return nil, res, nil
	}
	nodeEdpLister, err := resStore.NewLister(d.store, &core.Endpoint{}, &core.ResourceEventHandle{Interceptor: interceptor})
	if err != nil {
		return err
	}
	d.nodeEdpLister = nodeEdpLister
	return nil
}

func (d *daemon) cacheService() error {
	interceptor := func(put bool, r, prer core.Resource) (core.Resource, bool, error) {
		var nowCSvcs []*core.ContainerService
		if r != nil {
			nowCSvcs = d.parseContainerSvc(r.(*core.Service))
		}
		var preCSvcs []*core.ContainerService
		if prer != nil {
			preCSvcs = d.parseContainerSvc(prer.(*core.Service))
		}
		for _, nowCSvc := range nowCSvcs {
			isExist := false
			for _, preCSvc := range preCSvcs {
				if nowCSvc.Name == preCSvc.Name && preCSvc.Labels[core.HashLabelKey] == nowCSvc.Labels[core.HashLabelKey] {
					isExist = true
				}
			}
			if !isExist {
				d.svcCache.Store(nowCSvc.Name+"_"+nowCSvc.Labels[core.HashLabelKey], nowCSvc)
			}
		}
		for _, preCSvc := range preCSvcs {
			isExist := false
			for _, nowCSvc := range nowCSvcs {
				if nowCSvc.Name == preCSvc.Name && preCSvc.Labels[core.HashLabelKey] == nowCSvc.Labels[core.HashLabelKey] {
					isExist = true
				}
			}
			if !isExist {
				d.svcCache.Delete(preCSvc.Name + "_" + preCSvc.Labels[core.HashLabelKey])
			}
		}
		return nil, true, nil
	}
	handle := &core.ResourceEventHandle{
		Interceptor: interceptor,
	}
	svcLister, err := resStore.NewLister(d.store, &core.Service{}, handle)
	if err != nil {
		return err
	}
	d.svcLister = svcLister
	return nil
}

func (d *daemon) parseContainerSvc(svc *core.Service) []*core.ContainerService {
	cSvcs := make([]*core.ContainerService, 0)
	if svc.Kind != "docker" {
		return cSvcs
	}
	for _, ed := range svc.Endpoints {
		if ed.Hostname != d.node.Hostname {
			continue
		}
		if ed.Name == "" {
			ed.Name = ed.Hostname
		}
		ed.Domain = ed.Hostname + "." + svc.Name + "." + svc.Namespace
		vars := core.ServiceValues{
			Node: core.Node{
				Hostname: d.node.Hostname,
				IP:       d.node.IP,
			},
			Endpoint: ed,
		}
		container, err := svc.ParseContainer(vars)
		if err != nil {
			logrus.Error(err)
			continue
		}
		container.Name = d.containerName(svc, &ed)
		if container.Labels == nil {
			container.Labels = make(map[string]string)
		}
		container.Labels[core.HashLabelKey] = md5V(container)
		container.Labels[core.CreatorLabelKey] = "oars"
		cSvcs = append(cSvcs, container)
	}
	return cSvcs
}

func (d *daemon) cacheContainers() {
	t := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-t.C:
			//list docker containers and find container who status has updated
			cs, err := d.List(context.Background())
			if err != nil {
				logrus.Error(err)
				continue
			}
			edps := make(map[string]*core.Endpoint)
			putEps := make([]*core.Endpoint, 0)
			for _, cn := range cs {
				if _, ok := cn.Labels[core.CreatorLabelKey]; !ok {
					continue
				}
				edp := d.cantainerToEndpoint(cn)
				edps[edp.Status.ID] = edp
				if oldedp, ok := d.endpointCache[edp.Status.ID]; ok {
					if oldedp.Status.IP != edp.Status.IP || oldedp.Status.StateDetail != edp.Status.StateDetail {
						putEps = append(putEps, edp)
					}
					continue
				}
				putEps = append(putEps, edp)
			}
			d.mu.Lock()
			d.endpointCache = edps
			d.mu.Unlock()
			d.ready = true
			//find new add endpoints and create these
			d.svcCache.Range(func(k, v interface{}) bool {
				svc := v.(*core.ContainerService)
				edpExist := false
				for _, edp := range edps {
					if svc.Name == d.containerNameByEdp(edp) {
						edpExist = true
						break
					}
				}
				if !edpExist {
					putEps = append(putEps, d.cserviceToEndpoint(svc))
				}
				return true
			})
			//update endpoints status
			for _, edp := range putEps {
				_, err = d.edpstore.Put(context.Background(), edp, &core.PutOptions{})
				if err != nil {
					logrus.Error(err)
				}
			}
			//gc endpoints that service had deleted
			resources, _ := d.nodeEdpLister.List()
			for _, resource := range resources {
				endpoint := resource.(*core.Endpoint)
				edpExist := false
				d.svcCache.Range(func(k, v interface{}) bool {
					svc := v.(*core.ContainerService)
					if svc.Name == d.containerNameByEdp(endpoint) {
						edpExist = true
						return false
					}
					return true
				})
				if !edpExist {
					for _, edp := range edps {
						if d.containerNameByEdp(endpoint) == d.containerNameByEdp(edp) {
							edpExist = true
							break
						}
					}
				}
				if !edpExist {
					err = d.edpstore.Delete(context.Background(), endpoint, &core.DeleteOptions{})
					if err != nil {
						logrus.Error(err)
					}
				}
			}
		}
	}
}

func (d *daemon) syncDockerSvc() error {
	d.mu.Lock()
	svcs := make([]*core.ContainerService, 0)
	d.svcCache.Range(func(k, v interface{}) bool {
		svcs = append(svcs, v.(*core.ContainerService))
		return true
	})
	delList := make([]*core.Endpoint, 0)
	for _, edp := range d.endpointCache {
		edpExist := false
		for _, svc := range svcs {
			if svc.Name == d.containerNameByEdp(edp) && svc.Labels[core.HashLabelKey] == edp.Labels[core.HashLabelKey] {
				edpExist = true
				break
			}

		}
		if !edpExist {
			delList = append(delList, edp)
		}
	}
	addList := make([]*core.ContainerService, 0)
	for _, svc := range svcs {
		edpExist := false
		for _, edp := range d.endpointCache {
			if svc.Name == d.containerNameByEdp(edp) && svc.Labels[core.HashLabelKey] == edp.Labels[core.HashLabelKey] {
				edpExist = true
				continue
			}
		}
		if !edpExist {
			addList = append(addList, svc)
		}
	}
	d.mu.Unlock()
	ignoreCreates := make([]*core.Endpoint, 0)
	ctx := context.Background()
	//TODO 并发？
	//删除容器
	for _, endpoint := range delList {
		err := d.Remove(ctx, endpoint.Status.ID)
		if err != nil {
			if d.dockerError(err) != errNotFound {
				logrus.Error(err)
				ignoreCreates = append(ignoreCreates, endpoint)
				continue
			}
		}
	}

	//创建容器
	for _, svc := range addList {
		for _, oldendpoint := range ignoreCreates {
			if svc.Name == d.containerNameByEdp(oldendpoint) {
				continue
			}
		}
		err := d.Create(ctx, svc)
		if err != nil {
			logrus.Error(err)
			continue
		}
		//防止新建容器未同步，导致重复创建
		edp := d.getEndpointByContainerName(svc.Name)
		edp.Labels = svc.Labels
		d.mu.Lock()
		d.endpointCache[d.containerNameByEdp(edp)] = edp
		d.mu.Unlock()
	}
	return nil
}
