package admin

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

func (s *service) initConfigMap() {
	cfg := &core.ConfigMap{
		ResourceMeta: &core.ResourceMeta{
			Name:      core.SystemConfigName,
			Namespace: core.SystemNamespace,
		},
		Data: make(map[string]string),
	}
	_, err := s.cfgStore.Get(context.TODO(), cfg, &core.GetOptions{})
	if err == e.ErrResourceNotFound {
		s.cfgStore.Put(context.TODO(), cfg, &core.PutOptions{})
	}
}

func (s *service) regConfigMap(ctx context.Context, action string, args interface{}) *core.APIReply {
	switch action {
	case "get":
		return s.ListConfigMap(args)
	case "put":
		return s.PutConfigMap(args)
	case "delete":
		return s.DeleteConfigMap(args)
	}
	return e.MethodNotFoundMethod()
}

func (s *service) PutConfigMap(args interface{}) *core.APIReply {
	var cfg core.ConfigMap
	err := unmarshalArgs(args, &cfg)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	if !nameRegex.MatchString(cfg.Name) {
		return e.InvalidParameterError()
	}
	ctx := context.TODO()
	_, err = s.cfgStore.Put(ctx, &cfg, &core.PutOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(cfg)
}

func (s *service) DeleteConfigMap(args interface{}) *core.APIReply {
	var cfg core.ConfigMap
	err := unmarshalArgs(args, &cfg)
	if err != nil {
		return e.InternalError(err)
	}
	ctx := context.TODO()
	err = s.cfgStore.Delete(ctx, &cfg, &core.DeleteOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply("")
}

func (s *service) ListConfigMap(args interface{}) *core.APIReply {
	var cfg core.ConfigMap
	err := unmarshalArgs(args, &cfg)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	cfgs, err := s.cfgStore.List(ctx, &cfg, &core.ListOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(cfgs)
}
