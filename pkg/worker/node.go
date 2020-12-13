package worker

import (
	"context"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/worker/metrics"
	"github.com/sirupsen/logrus"
)

func (d *daemon) initNode() {
	d.putNode()
	t := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-t.C:
			d.putNode()
		}
	}

}

func (d *daemon) putNode() {
	nodeInfo, err := metrics.GetNodeInfo()
	if err != nil {
		logrus.Error(err)
	}
	endpoint := &core.Endpoint{
		Name:      d.node.Hostname,
		Namespace: "system",
		Service:   "node",
		Status: &core.EndpointStatus{
			ID:       d.node.Hostname,
			Port:     d.node.Port,
			IP:       d.node.IP,
			NodeInfo: nodeInfo,
			State:    "running",
			Node: core.Node{
				Hostname: d.node.Hostname,
				IP:       d.node.IP,
			},
		},
	}
	_, err = d.edpstore.Put(context.Background(), endpoint, &core.PutOptions{})
	if err != nil {
		logrus.Error(err)
	}
}
