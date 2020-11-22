package admin

import (
	"context"
	"errors"
	"fmt"
	"net/rpc"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

func (s *service) GetEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	key := fmt.Sprintf("services/endpoint/namespaces/%s/%s", endpoint.Namespace, endpoint.Service)
	if endpoint.Service != "" {
		key += "/"
	}
	if endpoint.Hostname != "" {
		key += endpoint.Hostname
	}
	ctx := context.TODO()
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}
	endpoints := make([]core.Endpoint, 0)
	for _, kv := range kvs {
		endpoint := new(core.Endpoint)
		endpoint.Parse(kv.Value)
		endpoints = append(endpoints, *endpoint)
	}
	return core.NewAPIReply(endpoints)
}

func (s *service) RestartEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	conn, err := s.getConn(endpoint.Hostname)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	err = conn.Call("Endpoint.EndpointRestart", endpoint, nil)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	return core.NewAPIReply("")
}
func (s *service) getConn(hostname string) (*rpc.Client, error) {
	r := s.GetEndPoint(core.Endpoint{
		Namespace: "system",
		Service:   "node",
		Hostname:  hostname,
	})
	if r.Code != core.ServiceSuccessCode {
		return nil, errors.New(r.SubCode)
	}
	res := r.Data.([]core.Endpoint)
	if len(res) == 0 {
		return nil, errors.New("not found node")
	}
	return rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", res[0].HostIP, res[0].Port))

}

func (s *service) StopEndPoint(args interface{}) *core.APIReply {
	var endpoint core.Endpoint
	err := unmarshalArgs(args, &endpoint)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	conn, err := s.getConn(endpoint.Hostname)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	err = conn.Call("Endpoint.EndpointStop", &endpoint, nil)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	return core.NewAPIReply("")
}

func (s *service) GetEndPointLog(args interface{}) *core.APIReply {
	var opt core.EndpointLogOpt
	err := unmarshalArgs(args, &opt)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	conn, err := s.getConn(opt.Hostname)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	var r core.APIReply
	err = conn.Call("Endpoint.EndpointLog", &opt, &r)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	return &r
}
