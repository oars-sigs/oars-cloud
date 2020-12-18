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
	if !nameRegex.MatchString(ns.Name) {
		return e.InvalidParameterError()
	}
	ctx := context.TODO()
	_, err = s.nsStore.Put(ctx, &ns, &core.PutOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(ns)
}

func (s *service) DeleteNamespace(args interface{}) *core.APIReply {
	var ns core.Namespace
	err := unmarshalArgs(args, &ns)
	if err != nil {
		return e.InternalError(err)
	}
	ctx := context.TODO()
	err = s.nsStore.Delete(ctx, &ns, &core.DeleteOptions{})
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
	nss, err := s.nsStore.List(ctx, &ns, &core.ListOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(nss)
}
