package admin

import (
	"context"
	"encoding/base64"
	"fmt"

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
	key := fmt.Sprintf("ingresses/listener/%s", listener.Name)
	ctx := context.TODO()
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}
	listeners := make([]core.IngressListener, 0)
	for _, kv := range kvs {
		lis := new(core.IngressListener)
		lis.Parse(kv.Value)
		listeners = append(listeners, *lis)
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

	v := core.KV{
		Key:   fmt.Sprintf("ingresses/listener/%s", listener.Name),
		Value: listener.String(),
	}
	ctx := context.TODO()
	err = s.store.Put(ctx, v)
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
	err = s.store.Delete(ctx, "ingresses/route/listener/"+listener.Name+"/", core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}

	err = s.store.Delete(ctx, "ingresses/listener/"+listener.Name, core.KVOption{WithPrefix: true})
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
	var ingress core.IngressRoute
	err := unmarshalArgs(args, &ingress)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	key := fmt.Sprintf("ingresses/route/listener/%s/namespaces/%s/%s", ingress.Namespace, ingress.Listener, ingress.Name)
	ctx := context.TODO()
	kvs, err := s.store.Get(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}
	ingresses := make([]core.IngressRoute, 0)
	for _, kv := range kvs {
		ing := new(core.IngressRoute)
		ing.Parse(kv.Value)
		ingresses = append(ingresses, *ing)
	}
	return core.NewAPIReply(ingresses)
}

func (s *service) PutIngressRoute(args interface{}) *core.APIReply {
	var ingress core.IngressRoute
	err := unmarshalArgs(args, &ingress)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	v := core.KV{
		Key:   fmt.Sprintf("ingresses/route/listener/%s/namespaces/%s/%s", ingress.Namespace, ingress.Listener, ingress.Name),
		Value: ingress.String(),
	}
	ctx := context.TODO()
	err = s.store.Put(ctx, v)
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(ingress)
}

func (s *service) DeleteIngressRoute(args interface{}) *core.APIReply {
	var ingress core.IngressRoute
	err := unmarshalArgs(args, &ingress)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	key := fmt.Sprintf("ingresses/route/listener/%s/namespaces/%s/%s", ingress.Namespace, ingress.Listener, ingress.Name)
	err = s.store.Delete(ctx, key, core.KVOption{WithPrefix: true})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply("")
}
