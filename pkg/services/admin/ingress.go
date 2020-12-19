package admin

import (
	"context"
	"encoding/base64"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/certificate"
	"github.com/oars-sigs/oars-cloud/pkg/e"
)

func (s *service) regIngressListener(ctx context.Context, action string, args interface{}) *core.APIReply {
	switch action {
	case "get":
		return s.GetIngressListener(args)
	case "put":
		return s.PutIngressListener(args)
	case "delete":
		return s.DeleteIngressListener(args)
	}
	return e.MethodNotFoundMethod()
}

func (s *service) GetIngressListener(args interface{}) *core.APIReply {
	var listener core.IngressListener
	err := unmarshalArgs(args, &listener)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	listeners, err := s.ingressListenerStore.List(ctx, &listener, &core.ListOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(listeners)
}

func (s *service) PutIngressListener(args interface{}) *core.APIReply {
	var listener core.IngressListener
	err := unmarshalArgs(args, &listener)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	if len(listener.TLSCerts) == 0 {
		ca := &certificate.CA{
			Name:   listener.Name,
			CN:     "oarscloud",
			Expiry: "87600h",
		}
		req := &certificate.Request{
			Name:      listener.Name,
			CN:        listener.Name,
			O:         listener.Name,
			Hostnames: []string{"oars"},
		}
		cert, err := certificate.Create(ca, req)
		if err != nil {
			return e.InternalError(err)
		}
		listener.TLSCerts = []core.TLSCertificate{
			core.TLSCertificate{
				Host: "*",
				Cert: base64.StdEncoding.EncodeToString([]byte(cert.Cert)),
				Key:  base64.StdEncoding.EncodeToString([]byte(cert.Key)),
			},
		}
	}
	ctx := context.TODO()
	_, err = s.ingressListenerStore.Put(ctx, &listener, &core.PutOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(listener)
}

func (s *service) DeleteIngressListener(args interface{}) *core.APIReply {
	var listener core.IngressListener
	err := unmarshalArgs(args, &listener)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	err = s.ingressListenerStore.Delete(ctx, &listener, &core.DeleteOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply("")
}

func (s *service) regIngressRoute(ctx context.Context, action string, args interface{}) *core.APIReply {
	switch action {
	case "get":
		return s.GetIngressRoute(args)
	case "put":
		return s.PutIngressRoute(args)
	case "delete":
		return s.DeleteIngressRoute(args)
	}
	return e.MethodNotFoundMethod()
}

func (s *service) GetIngressRoute(args interface{}) *core.APIReply {
	var route core.IngressRoute
	err := unmarshalArgs(args, &route)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	routes, err := s.ingressRouteStore.List(ctx, &route, &core.ListOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(routes)
}

func (s *service) PutIngressRoute(args interface{}) *core.APIReply {
	var route core.IngressRoute
	err := unmarshalArgs(args, &route)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	_, err = s.ingressRouteStore.Put(ctx, &route, &core.PutOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(route)
}

func (s *service) DeleteIngressRoute(args interface{}) *core.APIReply {
	var route core.IngressRoute
	err := unmarshalArgs(args, &route)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	err = s.ingressRouteStore.Delete(ctx, &route, &core.DeleteOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply("")
}
