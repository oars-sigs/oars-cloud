package agent

import (
	"sync"

	"github.com/docker/docker/client"

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
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	d := &daemon{
		c:     cli,
		store: store,
		node:  node,
	}
	go d.initNode()
	go d.run()
	go d.watch()
	go d.dnsServer()
	go metrics.Start(cli, node)
	err = d.reg()
	return err
}
