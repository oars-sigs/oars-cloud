package worker

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	resStore "github.com/oars-sigs/oars-cloud/pkg/store/resources"
	"github.com/oars-sigs/oars-cloud/pkg/utils/strvars"
	"github.com/sirupsen/logrus"
)

func (d *daemon) run() {
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

func (d *daemon) cacheConfig() error {
	interceptor := func(put bool, r, prer core.Resource) (core.Resource, bool, error) {
		if put {
			cfg := r.(*core.ConfigMap)
			if cfg.Name == core.DefaultSystemName && cfg.Namespace == core.SystemNamespace {
				systemCfg := new(core.SystemConfig)
				err := strvars.Parse(cfg.Data, systemCfg)
				if err != nil {
					logrus.Error(err)
					return nil, true, nil
				}
				d.sysConfig = systemCfg
			}
		}
		return nil, true, nil
	}
	cfgLister, err := resStore.NewLister(d.store, new(core.ConfigMap), &core.ResourceEventHandle{Interceptor: interceptor})
	if err != nil {
		return err
	}
	d.cfgLister = cfgLister
	return nil
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
	ename := ""
	enames := strings.Split(svc.Name, "@")
	if len(enames) > 1 {
		svc.Name = enames[0]
		ename = enames[1]
	}
	for _, ed := range svc.Endpoints {
		if ed.Hostname != d.node.Hostname {
			continue
		}
		if ed.Name == "" {
			if ename != "" {
				ed.Name = ename
			} else {
				ed.Name = ed.Hostname
			}
		}
		ed.Domain = ed.Name + "." + svc.Name + "." + svc.Namespace
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
		if container.NetworkMode == "" {
			container.NetworkMode = d.node.ContainerNetwork
		}
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
					if oldedp.Status.IP != edp.Status.IP || oldedp.Status.State != edp.Status.State || oldedp.Status.ID != edp.Status.ID {
						edp.SetCreated(time.Now().Unix())
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
					d.delEvent(endpoint, "", "")
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
			if svc.Name == d.containerNameByEdp(edp) {
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
			if svc.Name == d.containerNameByEdp(edp) {
				if svc.Labels[core.HashLabelKey] != edp.Labels[core.HashLabelKey] {
					svc.ID = edp.Status.ID
					break
				}
				edpExist = true
				break
			}
		}
		if !edpExist {
			addList = append(addList, svc)
		}
	}
	d.mu.Unlock()
	ctx := context.Background()

	//删除容器
	delGw := new(sync.WaitGroup)
	delGw.Add(len(delList))
	for _, endpoint := range delList {
		go func(edp *core.Endpoint) {
			d.addEvent(edp, core.DeleteEventAction, core.InProgressEventStatus, "")
			err := d.Remove(ctx, edp.Status.ID)
			if err != nil {
				if d.dockerError(err) != errNotFound {
					logrus.Error(err)
					d.addEvent(edp, core.DeleteEventAction, core.FailEventStatus, err.Error())
				}
			}
			d.addEvent(edp, core.DeleteEventAction, core.SuccessEventStatus, "")
			delGw.Done()
		}(endpoint)
	}

	//TODO 并发？
	//创建容器
	for _, svc := range addList {
		//如果有旧容器，先删除
		edp := d.cserviceToEndpoint(svc)
		if svc.ID != "" {
			d.addEvent(edp, core.DeleteEventAction, core.InProgressEventStatus, "")
			err := d.Remove(ctx, svc.ID)
			if err != nil {
				if d.dockerError(err) != errNotFound {
					logrus.Error(err)
					d.addEvent(edp, core.DeleteEventAction, core.FailEventStatus, err.Error())
					continue
				}
			}
			d.addEvent(edp, core.DeleteEventAction, core.SuccessEventStatus, "")
		}

		//vault
		if d.vault != nil {
			for i, v := range svc.Environment {
				kv := strings.Split(v, "=")
				if len(kv) < 2 {
					continue
				}
				if strings.HasPrefix(kv[1], "$oars_vault:") {
					keys := strings.Split(kv[1], ":")
					if len(keys) != 3 {
						continue
					}
					value, _ := d.vault.Get(keys[1], keys[2])
					svc.Environment[i] = kv[0] + "=" + value
				}
			}
		}

		//registry auth
		if svc.ImagePullAuth == "" && d.sysConfig != nil {
			for _, registry := range d.sysConfig.ImageRegistry {
				if strings.HasPrefix(svc.Image, registry.Address+"/") {
					svc.ImagePullAuth = registryAuth(registry.Username, registry.Password)
				}
			}
		}
		//pull image
		d.addEvent(edp, core.ImagePullEventAction, core.InProgressEventStatus, "")
		err := d.ImagePull(ctx, svc)
		if err != nil {
			logrus.Error(err)
			d.addEvent(edp, core.ImagePullEventAction, core.FailEventStatus, err.Error())
			continue
		}
		d.addEvent(edp, core.ImagePullEventAction, core.SuccessEventStatus, "")

		//create
		d.addEvent(edp, core.CreateEventAction, core.InProgressEventStatus, "")
		id, err := d.Create(ctx, svc)
		if err != nil {
			logrus.Error(err)
			d.addEvent(edp, core.CreateEventAction, core.FailEventStatus, err.Error())
			continue
		}
		d.addEvent(edp, core.CreateEventAction, core.SuccessEventStatus, "")
		//start
		go func() {
			d.addEvent(edp, core.StartEventAction, core.InProgressEventStatus, "")
			err = d.Start(ctx, id)
			if err != nil {
				d.addEvent(edp, core.StartEventAction, core.FailEventStatus, err.Error())
				logrus.Error(err)
				return
			}
			d.addEvent(edp, core.StartEventAction, core.SuccessEventStatus, "")
		}()
		//防止新建容器未同步，导致重复创建
		d.mu.Lock()
		d.endpointCache[d.containerNameByEdp(edp)] = edp
		d.mu.Unlock()
	}
	delGw.Wait()
	return nil
}
