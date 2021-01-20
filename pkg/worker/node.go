package worker

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	resStore "github.com/oars-sigs/oars-cloud/pkg/store/resources"
	"github.com/oars-sigs/oars-cloud/pkg/worker/metrics"

	"github.com/sirupsen/logrus"
)

func (d *daemon) initNode() error {
	nodeInfo, err := metrics.GetNodeInfo()
	if err != nil {
		return err
	}
	endpoint := &core.Endpoint{
		ResourceMeta: &core.ResourceMeta{
			Name:      d.node.Hostname,
			Namespace: "system",
		},
		Kind:    "runtime",
		Service: "node",
		Status: &core.EndpointStatus{
			ID:       d.node.Hostname,
			Port:     d.node.Port,
			IP:       d.node.IP,
			NodeInfo: nodeInfo,
			State:    "running",
			Node: core.Node{
				Hostname:      d.node.Hostname,
				IP:            d.node.IP,
				ContainerCIDR: d.node.ContainerCIDR,
			},
		},
	}
	_, err = d.edpstore.Put(context.Background(), endpoint, &core.PutOptions{})
	if err != nil {
		return err
	}
	endpoint.ResourceMeta.ObjectKind = &core.ResourceObjectKind{
		IsRegister: true,
	}
	_, err = resStore.NewRegister(d.store, endpoint, 10)
	if err != nil {
		return err
	}
	if d.node.ContainerCIDR != "" {
		d.node.Interface = d.getInterface()
		//config network
		go d.configNetwork()
	}
	return err
}

func (d *daemon) configNetwork() {
	t := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-t.C:
			ress, ok := d.edpLister.List()
			if !ok {
				return
			}
			cidrs := make([]string, 0)
			for _, res := range ress {
				edp := res.(*core.Endpoint)
				if edp.Service == "node" && edp.Namespace == "system" &&
					edp.Name != d.node.Hostname && edp.Status.Node.ContainerCIDR != "" {
					cidrs = append(cidrs, edp.Status.Node.ContainerCIDR)
				}
			}
			err := reconcileRouters(d.node.Interface, cidrs)
			if err != nil {
				logrus.Error(err)
			}
		}
	}
}

func (d *daemon) addEvent(r core.Resource, action, status, msg string) {
	d.delEvent(r, action, status)
	event := d.convEvent(r, action, status, msg)
	_, err := d.eventstore.Put(context.Background(), event, &core.PutOptions{})
	if err != nil {
		logrus.Error(err)
	}
}

func (d *daemon) delEvent(r core.Resource, action, status string) {
	event := d.convEvent(r, action, status, "")
	err := d.eventstore.Delete(context.Background(), event, &core.DeleteOptions{})
	if err != nil {
		logrus.Error(err)
	}
}
