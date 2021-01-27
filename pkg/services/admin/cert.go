package admin

import (
	"context"
	"encoding/base64"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
	"github.com/oars-sigs/oars-cloud/pkg/utils/rsa"
)

func (s *service) regCert(ctx context.Context, action string, args interface{}) *core.APIReply {
	switch action {
	case "put":
		return s.PutCert(args)
	case "get":
		return s.GetCert(args)
	case "delete":
		return s.DeleteCert(args)
	}
	return e.MethodNotFoundMethod()
}

func (s *service) initCert() {
	rootCert := &core.Certificate{
		ResourceMeta: &core.ResourceMeta{
			Name: core.DefaultRootCertName,
		},
		Info: &core.CertInformation{
			CommonName: "Oarscloud",
			Expires:    100,
			IsCA:       true,
		},
	}
	_, err := s.certStore.Get(context.TODO(), rootCert, &core.GetOptions{})
	if err == e.ErrResourceNotFound {
		s.PutCert(rootCert)
	}
	defaultCert := &core.Certificate{
		ResourceMeta: &core.ResourceMeta{
			Name: core.DefaultServerCertName,
		},
		RootCA: core.DefaultRootCertName,
		Info: &core.CertInformation{
			CommonName: "OarsCloudServer",
			Expires:    100,
		},
	}
	_, err = s.certStore.Get(context.TODO(), defaultCert, &core.GetOptions{})
	if err == e.ErrResourceNotFound {
		s.PutCert(defaultCert)
	}
}

func (s *service) PutCert(args interface{}) *core.APIReply {
	var cert core.Certificate
	err := unmarshalArgs(args, &cert)
	if err != nil {
		return e.InternalError(err)
	}
	if cert.Cert == "" {
		if cert.Info.IsCA {
			crt, key, err := rsa.CreateCRT(nil, nil, cert.Info)
			if err != nil {
				return e.InternalError(err)
			}
			cert.Cert = base64.StdEncoding.EncodeToString(crt)
			cert.Key = base64.StdEncoding.EncodeToString(key)
			p12, _ := rsa.CertToP12(crt, key, "")
			cert.P12 = string(p12)
		}
		if !cert.Info.IsCA {
			if cert.RootCA == "" {
				cert.RootCA = core.DefaultRootCertName
			}
			reply := s.GetCert(&core.Certificate{ResourceMeta: &core.ResourceMeta{Name: cert.RootCA}})
			if reply.Code != core.ServiceSuccessCode {
				return reply
			}
			rootCerts := reply.Data.([]core.Resource)
			if len(rootCerts) == 0 {
				return e.InvalidParameterError(e.ErrCACertNotFound)
			}
			rootCert := rootCerts[0].(*core.Certificate)
			newcrt, err := base64.StdEncoding.DecodeString(rootCert.Cert)
			if err != nil {
				return e.InternalError(err)
			}
			newkey, err := base64.StdEncoding.DecodeString(rootCert.Key)
			if err != nil {
				return e.InternalError(err)
			}
			rootCrt, err := rsa.ParseCrt(newcrt)
			if err != nil {
				return e.InternalError(err)
			}
			rootKey, err := rsa.ParseKey(newkey)
			if err != nil {
				return e.InternalError(err)
			}
			crt, key, err := rsa.CreateCRT(rootCrt, rootKey, cert.Info)
			if err != nil {
				return e.InternalError(err)
			}
			cert.Cert = base64.StdEncoding.EncodeToString(crt)
			cert.Key = base64.StdEncoding.EncodeToString(key)
			p12, _ := rsa.CertToP12(crt, key, "")
			cert.P12 = string(p12)
		}
	}
	if cert.Info == nil {
		crt, err := base64.StdEncoding.DecodeString(cert.Cert)
		if err != nil {
			return e.InternalError(err)
		}
		newcert, err := rsa.ParseCrt(crt)
		if err != nil {
			return e.InternalError(err)
		}
		cert.Info = rsa.ParseCertToInfo(newcert)
	}
	_, err = s.certStore.Put(context.TODO(), &cert, &core.PutOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(cert)
}

func (s *service) GetCert(args interface{}) *core.APIReply {
	var cert core.Certificate
	err := unmarshalArgs(args, &cert)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	certs, err := s.certStore.List(ctx, &cert, &core.ListOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply(certs)
}

func (s *service) DeleteCert(args interface{}) *core.APIReply {
	var cert core.Certificate
	err := unmarshalArgs(args, &cert)
	if err != nil {
		return e.InvalidParameterError(err)
	}
	ctx := context.TODO()
	err = s.certStore.Delete(ctx, &cert, &core.DeleteOptions{})
	if err != nil {
		return e.InternalError(err)
	}
	return core.NewAPIReply("")
}
