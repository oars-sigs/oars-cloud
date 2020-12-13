package worker

import (
	"sync"

	"github.com/docker/docker/client"

	"github.com/oars-sigs/oars-cloud/core"
	edpStore "github.com/oars-sigs/oars-cloud/pkg/store/endpoints"
	"github.com/oars-sigs/oars-cloud/pkg/worker/metrics"
)

type daemon struct {
	c             *client.Client
	store         core.KVStore
	svcLister     core.ResourceLister
	edpLister     core.ResourceLister
	edpstore      core.EndpointStore
	mu            *sync.Mutex
	endpointCache map[string]*core.Endpoint //current node endpoints
	svcCache      sync.Map                  //current node services
	node          core.NodeConfig
	ready         bool
}

//Start ...
func Start(store core.KVStore, node core.NodeConfig) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	edpstore := edpStore.New(store)
	d := &daemon{
		c:             cli,
		store:         store,
		node:          node,
		mu:            new(sync.Mutex),
		endpointCache: make(map[string]*core.Endpoint),
		edpstore:      edpstore,
	}
	go d.initNode()
	go d.run()
	go d.dnsServer()
	go metrics.Start(cli, node)
	err = d.reg()
	return err
}
