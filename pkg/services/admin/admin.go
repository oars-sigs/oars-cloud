package admin

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

type service struct {
	store core.KVStore
}

//New admin api
func New(store core.KVStore) core.ServiceInterface {
	s := &service{store}
	s.PutNamespace(core.Namespace{Name: "system"})
	s.PutService(core.Service{Namespace: "system", Name: "admin", Kind: "runtime"})
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
	case "ingress":
		r = s.regIngress(ctx, action, args)
	default:
		r = e.MethodNotFoundMethod()
	}
	if reply != nil {
		*reply = *r
	}

	return nil
}
