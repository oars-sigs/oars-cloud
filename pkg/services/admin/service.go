package admin

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

func (s *service) regService(ctx context.Context, action string, args interface{}) *core.APIReply {
	switch action {
	case "get":
		return s.GetService(args)
	case "put":
		return s.PutService(args)
	case "delete":
		return s.DeleteService(args)
	}
	return e.MethodNotFoundMethod()
}

func (s *service) PutService(args interface{}) *core.APIReply {
	var svc core.Service
	err := unmarshalArgs(args, &svc)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	if !nameRegex.MatchString(svc.Name) {
		return e.InvalidParameterError()
	}
	ctx := context.TODO()
	v := core.KV{
		Key:   "services/svc/namespaces/" + svc.Namespace + "/" + svc.Name,
		Value: svc.String(),
	}
	err = s.store.Put(ctx, v)
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(svc)
}

func (s *service) DeleteService(args interface{}) *core.APIReply {
	var svc core.Service
	err := unmarshalArgs(args, &svc)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	err = s.store.Delete(ctx, "services/method/namespaces/"+svc.Namespace+"/"+svc.Name+"/", core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}
	err = s.store.Delete(ctx, "services/svc/namespaces/"+svc.Namespace+"/"+svc.Name, core.KVOption{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply("")
}

func (s *service) GetService(args interface{}) *core.APIReply {
	var svc core.Service
	err := unmarshalArgs(args, &svc)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	kvs, err := s.store.Get(ctx, "services/svc/namespaces/"+svc.Namespace+"/"+svc.Name, core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}
	svcs := make([]core.Service, 0)
	for _, kv := range kvs {
		svc := new(core.Service)
		svc.Parse(kv.Value)
		svcs = append(svcs, *svc)
	}
	return core.NewAPIReply(svcs)
}
