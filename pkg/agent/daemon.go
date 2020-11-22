package agent

import (
	"os"
	"sync"

	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/agent/metrics"
)

type daemon struct {
	c             *client.Client
	store         core.KVStore
	dockerSvc     sync.Map
	endpointCache sync.Map
	node          core.NodeConfig
	ready         bool
}

//Start ...
func Start(store core.KVStore, node core.NodeConfig) error {
	os.Setenv("DOCKER_HOST", "tcp://127.0.0.1:2375")
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	d := &daemon{
		c:     cli,
		store: store,
		node:  node,
	}
	d.initNode()
	go d.run()
	go d.watch()
	go d.dnsServer()
	go metrics.Start(cli, node)
	err = d.reg()
	return err
}

func (d *daemon) initNode() {
	nodeInfo, err := metrics.GetNodeInfo()
	if err != nil {
		logrus.Error(err)
	}
	endpoint := &core.Endpoint{
		Namespace: "system",
		Service:   "node",
		Status:    "running",
		Hostname:  d.node.Hostname,
		HostIP:    d.node.IP,
		Port:      d.node.Port,
		NodeInfo:  nodeInfo,
	}
	err = d.putEndPoint(endpoint)
	if err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
}
