package admin

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

func (s *service) regMethod(ctx context.Context, action string, args interface{}) *core.APIReply {
	switch action {
	case "get":
		return s.GetMethod(args)
	case "put":
		return s.PutMethod(args)
	case "delete":
		return s.DeleteMethod(args)
	}
	return e.MethodNotFoundMethod()
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
