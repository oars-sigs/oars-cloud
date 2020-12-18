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
	_, err = s.svcStore.Put(ctx, &svc, &core.PutOptions{})
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
	err = s.svcStore.Delete(ctx, &svc, &core.DeleteOptions{})
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
	svcs, err := s.svcStore.List(ctx, &svc, &core.ListOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(svcs)
}
