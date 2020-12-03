package admin

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

func (s *service) regNamespace(ctx context.Context, action string, args interface{}) *core.APIReply {
	switch action {
	case "get":
		return s.GetNamespace(args)
	case "put":
		return s.PutNamespace(args)
	case "delete":
		return s.DeleteNamespace(args)
	}
	return e.MethodNotFoundMethod()
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
