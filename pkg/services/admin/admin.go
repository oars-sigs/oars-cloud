package admin

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
	"github.com/oars-sigs/oars-cloud/pkg/rpc"
	"github.com/oars-sigs/oars-cloud/pkg/store/resources"
)

type service struct {
	store                core.KVStore
	edpStore             core.ResourceStore
	nsStore              core.ResourceStore
	svcStore             core.ResourceStore
	ingressRouteStore    core.ResourceStore
	ingressListenerStore core.ResourceStore
	eventStore           core.ResourceStore
	certStore            core.ResourceStore
	cfgStore             core.ResourceStore
	cronStore            core.ResourceStore
	rpcClient            *rpc.Client
}

//New admin api
func New(store core.KVStore, rpcClient *rpc.Client, cfg *core.Config) core.ServiceInterface {
	s := &service{
		store:                store,
		edpStore:             resources.NewStore(store, new(core.Endpoint)),
		nsStore:              resources.NewStore(store, new(core.Namespace)),
		svcStore:             resources.NewStore(store, new(core.Service)),
		ingressRouteStore:    resources.NewStore(store, new(core.IngressRoute)),
		ingressListenerStore: resources.NewStore(store, new(core.IngressListener)),
		eventStore:           resources.NewStore(store, new(core.Event)),
		certStore:            resources.NewStore(store, new(core.Certificate)),
		cfgStore:             resources.NewStore(store, new(core.ConfigMap)),
		cronStore:            resources.NewStore(store, new(core.Cron)),
		rpcClient:            rpcClient,
	}
	s.PutNamespace(core.Namespace{
		ResourceMeta: &core.ResourceMeta{
			Name: "system",
		},
	})
	s.PutService(core.Service{
		ResourceMeta: &core.ResourceMeta{
			Name:      "admin",
			Namespace: "system",
		},
		Kind: "runtime",
	})
	s.PutService(core.Service{
		ResourceMeta: &core.ResourceMeta{
			Name:      "node",
			Namespace: "system",
		},
		Kind: "runtime",
	})
	s.edpStore.Put(context.Background(), &core.Endpoint{
		ResourceMeta: &core.ResourceMeta{
			Name:      cfg.Server.Name,
			Namespace: "system",
		},
		Service: "admin",
		Kind:    "runtime",
		Status: &core.EndpointStatus{
			ID:    cfg.Server.Name,
			IP:    cfg.Server.Host,
			State: "running",
		},
	}, &core.PutOptions{})
	s.initCert()
	s.initConfigMap()
	return s
}

func (s *service) Call(ctx context.Context, resource, action string, args interface{}, reply *core.APIReply) error {
	var r *core.APIReply
	switch resource {
	case "namespace":
		r = s.regNamespace(ctx, action, args)
	case "service":
		r = s.regService(ctx, action, args)
	case "endpoint":
		r = s.regEndpoint(ctx, action, args)
	case "ingressListener":
		r = s.regIngressListener(ctx, action, args)
	case "ingressRoute":
		r = s.regIngressRoute(ctx, action, args)
	case "event":
		r = s.regEvent(ctx, action, args)
	case "util":
		r = s.regUtil(ctx, action, args)
	case "cert":
		r = s.regCert(ctx, action, args)
	case "configmap":
		r = s.regConfigMap(ctx, action, args)
	case "cron":
		r = s.regCron(ctx, action, args)
	default:
		r = e.ResourceNotFoundError()
	}
	if reply != nil {
		*reply = *r
	}

	return nil
}
