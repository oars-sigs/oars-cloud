package agent

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/sirupsen/logrus"
)

func (d *daemon) watch() error {
	kvs, rev, err := d.store.GetWithRev(context.TODO(), "services/", core.KVOption{WithPrefix: true})
	if err != nil {
		return err
	}
	for _, kv := range kvs {
		d.actionWatch(kv, true)
	}
	d.ready = true
	updateCh := make(chan core.WatchChan)
	errCh := make(chan error)
	go d.store.Watch(context.TODO(), "services/", updateCh, errCh, core.KVOption{DisableFirst: true, WithRev: rev, WithPrevKV: true, WithPrefix: true})
	for {
		select {
		case res := <-updateCh:
			kv := res.KV
			if !res.Put {
				kv = res.PrevKV
			}
			d.actionWatch(kv, res.Put)

		case err := <-errCh:
			fmt.Println(err)
		}
	}
}

func (d *daemon) actionWatch(kv core.KV, put bool) {

	if strings.HasPrefix(kv.Key, "services/svc") {
		svc := new(core.Service)
		svc.Parse(kv.Value)

		if put {
			match := false
			for _, n := range svc.Endpoints {
				if d.node.Hostname == n.Hostname {
					match = true
				}
			}
			if !match {
				d.dockerSvc.Delete(d.serviceName(svc))
				return
			}
			switch svc.Kind {
			case "docker":
				svc.ParseTpl(d.node.Hostname, nil)
				svc.Docker.Name = d.containerName(svc)
				d.dockerSvc.Store(d.serviceName(svc), svc)
			}
			return
		}
		d.dockerSvc.Delete(d.serviceName(svc))
		return
	}

	if strings.HasPrefix(kv.Key, "services/endpoint") {
		endpoint := new(core.Endpoint)
		endpoint.Parse(kv.Value)
		if put {
			d.endpointCache.Store(d.getCacheEndpointKey(endpoint), endpoint)
		}
		if !put {
			d.endpointCache.Delete(d.getCacheEndpointKey(endpoint))
		}
	}
}

func (d *daemon) run() {
	t := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-t.C:
			err := d.syncDockerSvc()
			if err != nil {
				logrus.Error(err)
			}
		}
	}
}

func (d *daemon) syncDockerSvc() error {
	if !d.ready {
		return nil
	}
	ctx := context.Background()
	cs, err := d.List(ctx)
	if err != nil {
		return err
	}

	delList := make([]string, 0)
	for _, cn := range cs {
		if !strings.HasPrefix(cn.Names[0], "/oars_") {
			continue
		}

		cname := strings.TrimPrefix(cn.Names[0], "/")
		exist := false
		var svc *core.Service
		d.dockerSvc.Range(func(key interface{}, v interface{}) bool {
			svc = v.(*core.Service)
			if cname == svc.Docker.Name {
				exist = true
			}
			return true
		})
		if !exist {
			delList = append(delList, cn.ID)
			//删除endpoint
			endpoint := d.getEndpointByContainerName(cname)
			err := d.deleteEndPoint(endpoint)
			if err != nil {
				logrus.Error(err)
			}
		}

		if exist {
			//更新endpoint
			endpoint := d.getEndpointByContainerName(cname)
			v, ok := d.endpointCache.Load(d.getCacheEndpointKey(endpoint))
			putFlag := false
			if !ok {
				putFlag = true
			}
			if ok {
				oldEndpoint := v.(*core.Endpoint)
				putFlag = (oldEndpoint.Status != cn.Status)
			}
			if putFlag {
				portStr, _ := cn.Labels["servicePort"]
				port, _ := strconv.Atoi(portStr)
				endpoint.Port = port
				endpoint.State = cn.State
				endpoint.Created = cn.Created
				endpoint.Status = cn.Status
				endpoint.ID = cn.ID
				err = d.putEndPoint(endpoint)
				if err != nil {
					logrus.Error(err)
				}
			}
		}

	}
	addList := make([]*core.Service, 0)
	d.dockerSvc.Range(func(key interface{}, v interface{}) bool {
		svc := v.(*core.Service)
		exist := false
		for _, cn := range cs {
			if strings.TrimPrefix(cn.Names[0], "/") == svc.Docker.Name {
				exist = true
			}
		}
		if !exist {
			addList = append(addList, svc)
		}
		return true
	})

	//删除容器
	for _, id := range delList {
		err := d.Remove(ctx, id)
		if err != nil {
			logrus.Error(err)
		}
	}

	//创建容器
	for _, svc := range addList {
		endpoint := d.getEndpointByContainerName(d.containerName(svc))
		endpoint.State = "creating"
		d.putEndPoint(endpoint)
		err := d.Create(ctx, svc)
		if err != nil {
			logrus.Error(err)
			endpoint.State = "error"
			endpoint.Status = err.Error()
			err = d.putEndPoint(endpoint)
			continue
		}
	}
	return nil

}
