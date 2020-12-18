package admin

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
	"github.com/oars-sigs/oars-cloud/pkg/store/endpoints"
	"github.com/oars-sigs/oars-cloud/pkg/store/ingresses"
	"github.com/oars-sigs/oars-cloud/pkg/store/namespaces"
	"github.com/oars-sigs/oars-cloud/pkg/store/services"
)

type service struct {
	store                core.KVStore
	edpStore             core.EndpointStore
	nsStore              core.NamespaceStore
	svcStore             core.ServiceStore
	ingressRouteStore    core.IngressRouteStore
	ingressListenerStore core.IngressListenerStore
}

//New admin api
func New(store core.KVStore, cfg *core.Config) core.ServiceInterface {
	s := &service{
		store:                store,
		edpStore:             endpoints.New(store),
		nsStore:              namespaces.New(store),
		svcStore:             services.New(store),
		ingressRouteStore:    ingresses.NewRoute(store),
		ingressListenerStore: ingresses.NewListener(store),
	}
	s.PutNamespace(core.Namespace{Name: "system"})
	s.PutService(core.Service{Namespace: "system", Name: "admin", Kind: "runtime"})
	s.PutService(core.Service{Namespace: "system", Name: "node", Kind: "runtime"})
	s.edpStore.Put(context.Background(), &core.Endpoint{
		Name:      cfg.Server.Name,
		Namespace: "system",
		Service:   "admin",
		Status: &core.EndpointStatus{
			ID:    cfg.Server.Name,
			IP:    cfg.Server.Host,
			State: "running",
		},
	}, &core.PutOptions{})
	return s
}

func (s *service) Call(ctx context.Context, resource, action string, args interface{}, reply *core.APIReply) error {
	var r *core.APIReply
	switch resource {
	case "namespace":
		r = s.regNamespace(ctx, action, args)
	case "service":
		r = s.regService(ctx, action, args)
	case "method":
		r = s.regService(ctx, action, args)
	case "endpoint":
		r = s.regEndpoint(ctx, action, args)
	case "ingressListener":
		r = s.regIngressListener(ctx, action, args)
	case "ingressRoute":
		r = s.regIngressRoute(ctx, action, args)
	default:
		r = e.MethodNotFoundMethod()
	}
	if reply != nil {
		*reply = *r
	}

	return nil
}
