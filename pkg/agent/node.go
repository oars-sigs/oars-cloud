package agent

import (
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/agent/metrics"
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
		ID:        d.node.Hostname,
		Namespace: "system",
		Service:   "node",
		State:     "running",
		Status:    "running",
		Hostname:  d.node.Hostname,
		HostIP:    d.node.IP,
		Port:      d.node.Port,
		NodeInfo:  nodeInfo,
		Updated:   time.Now().Unix(),
	}
	err = d.putEndPoint(endpoint)
	if err != nil {
		logrus.Error(err)
	}
}
