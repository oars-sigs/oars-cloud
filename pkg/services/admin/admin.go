package admin

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

type service struct {
	store core.KVStore
}

func New(store core.KVStore) core.ServiceInterface {
	s := &service{store}
	s.PutNamespace(core.Namespace{Name: "system"})
	s.PutService(core.Service{Namespace: "system", Name: "admin", Kind: "runtime"})
	return s
}

func (s *service) Call(ctx context.Context, resource, action string, args interface{}, reply *core.APIReply) error {
	var r *core.APIReply
	switch resource + "." + action {
	case "namespace.put":
		r = s.PutNamespace(args)
	case "service.put":
		r = s.PutService(args)
	case "method.put":
		r = s.PutMethod(args)
	case "namespace.delete":
		r = s.DeleteNamespace(args)
	case "service.delete":
		r = s.DeleteService(args)
	case "method.delete":
		r = s.DeleteMethod(args)
	case "namespace.get":
		r = s.GetNamespace(args)
	case "service.get":
		r = s.GetService(args)
	case "method.get":
		r = s.GetMethod(args)
	case "endpoint.get":
		r = s.GetEndPoint(args)
	case "endpoint.restart":
		r = s.RestartEndPoint(args)
	case "endpoint.stop":
		r = s.StopEndPoint(args)
	case "endpoint.log":
		r = s.GetEndPointLog(args)
	default:
		r = e.MethodNotFoundMethod()
	}
	if reply != nil {
		*reply = *r
	}

	return nil
}

func (s *service) PutNamespace(args interface{}) *core.APIReply {
	var ns core.Namespace
	err := unmarshalArgs(args, &ns)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	if ns.Name == "" {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	v := core.KV{
		Key:   "namespaces/" + ns.Name,
		Value: ns.String(),
	}
	err = s.store.Put(ctx, v)
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(ns)
}

func (s *service) DeleteNamespace(args interface{}) *core.APIReply {
	var ns core.Namespace
	err := unmarshalArgs(args, &ns)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	err = s.store.Delete(ctx, "services/method/namespaces/"+ns.Name+"/", core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}
	err = s.store.Delete(ctx, "services/svc/namespaces/"+ns.Name+"/", core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}

	err = s.store.Delete(ctx, "namespaces/"+ns.Name, core.KVOption{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply("")
}

func (s *service) GetNamespace(args interface{}) *core.APIReply {
	var ns core.Namespace
	err := unmarshalArgs(args, &ns)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	kvs, err := s.store.Get(ctx, "namespaces/"+ns.Name, core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}
	namespaces := make([]core.Namespace, 0)
	for _, kv := range kvs {
		ns := new(core.Namespace)
		ns.Parse(kv.Value)
		namespaces = append(namespaces, *ns)
	}
	return core.NewAPIReply(namespaces)
}

func (s *service) PutService(args interface{}) *core.APIReply {
	var svc core.Service
	err := unmarshalArgs(args, &svc)
	if err != nil {
		return e.InvalidParameterError(err)
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

func (s *service) PutMethod(args interface{}) *core.APIReply {
	var m core.Method
	err := unmarshalArgs(args, &m)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	v := core.KV{
		Key:   "services/method/namespaces/" + m.Namespace + "/" + m.ServiceName + "/" + m.Name,
		Value: m.String(),
	}
	err = s.store.Put(ctx, v)
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(m)
}

func (s *service) DeleteMethod(args interface{}) *core.APIReply {
	var m core.Method
	err := unmarshalArgs(args, &m)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	err = s.store.Delete(ctx, "services/method/namespaces/"+m.Namespace+"/"+m.ServiceName+"/"+m.Name, core.KVOption{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply("")
}

func (s *service) GetMethod(args interface{}) *core.APIReply {
	var m core.Method
	err := unmarshalArgs(args, &m)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	kvs, err := s.store.Get(ctx, "services/method/namespaces/"+m.Namespace+"/"+m.ServiceName+"/"+m.Name, core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}
	ms := make([]core.Method, 0)
	for _, kv := range kvs {
		m := new(core.Method)
		m.Parse(kv.Value)
		ms = append(ms, *m)
	}
	return core.NewAPIReply(ms)
}
