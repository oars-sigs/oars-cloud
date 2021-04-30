package worker

import (
	"sync"

	"github.com/docker/docker/client"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/cruntime/docker"
	"github.com/oars-sigs/oars-cloud/pkg/e"
	resStore "github.com/oars-sigs/oars-cloud/pkg/store/resources"
	"github.com/oars-sigs/oars-cloud/pkg/worker/metrics"
)

type daemon struct {
	c             *client.Client
	cri           core.ContainerRuntimeInterface
	store         core.KVStore
	svcLister     core.ResourceLister
	edpLister     core.ResourceLister
	nodeEdpLister core.ResourceLister
	cfgLister     core.ResourceLister
	edpstore      core.ResourceStore
	eventstore    core.ResourceStore
	mu            *sync.Mutex
	endpointCache map[string]*core.Endpoint //current node endpoints
	svcCache      sync.Map                  //current node services
	node          *core.NodeConfig
	sysConfig     *core.SystemConfig
	ready         bool
	vault         *VaultClient
}

//Start ...
func Start(store core.KVStore, node core.NodeConfig) error {
	var cri core.ContainerRuntimeInterface
	var err error
	switch node.ContainerDRIVE {
	case "docker":
		cri, err = docker.New()
		if err != nil {
			return err
		}
	default:
		return e.ErrInvalidContainerDrive
	}
	edpstore := resStore.NewStore(store, new(core.Endpoint))
	eventstore := resStore.NewStore(store, new(core.Event))
	d := &daemon{
		cri:           cri,
		store:         store,
		node:          &node,
		mu:            new(sync.Mutex),
		endpointCache: make(map[string]*core.Endpoint),
		edpstore:      edpstore,
		eventstore:    eventstore,
	}
	if node.Vault.Address != "" {
		c, err := newVault(node.Vault.Address, node.Vault.TOKEN)
		if err != nil {
			return err
		}
		d.vault = c
	}
	err = d.cacheConfig()
	if err != nil {
		return err
	}
	err = d.cacheEndpoint()
	if err != nil {
		return err
	}
	err = d.cacheService()
	if err != nil {
		return err
	}
	err = d.initNode()
	if err != nil {
		return err
	}
	go d.run()
	go d.dnsServer()
	err = startLVS(d.svcLister, d.edpLister)
	if err != nil {
		return err
	}
	go metrics.Start(cri, node)
	err = d.reg()
	return err
}
