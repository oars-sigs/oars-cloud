package worker

import (
	"sync"

	"github.com/docker/docker/client"

	"github.com/oars-sigs/oars-cloud/core"
	resStore "github.com/oars-sigs/oars-cloud/pkg/store/resources"
	"github.com/oars-sigs/oars-cloud/pkg/worker/metrics"
)

type daemon struct {
	c             *client.Client
	store         core.KVStore
	svcLister     core.ResourceLister
	edpLister     core.ResourceLister
	nodeEdpLister core.ResourceLister
	edpstore      core.ResourceStore
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
	edpstore := resStore.NewStore(store, new(core.Endpoint))
	d := &daemon{
		c:             cli,
		store:         store,
		node:          node,
		mu:            new(sync.Mutex),
		endpointCache: make(map[string]*core.Endpoint),
		edpstore:      edpstore,
	}
	err = d.cacheEndpoint()
	if err != nil {
		return err
	}
	go d.initNode()
	go d.run()
	go d.dnsServer()
	go metrics.Start(cli, node)
	err = d.reg()
	return err
}
