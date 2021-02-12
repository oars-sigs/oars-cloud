package acme

import (
	"encoding/base64"
	"errors"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/utils/rsa"
)

// Client holds configurations of the provider.
type Client struct {
	client *lego.Client
	cert   *core.Certificate
}

//New new a provider
func New(cert *core.Certificate) (*Client, error) {
	config := lego.NewConfig(cert.Acme.Account)
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	err = setProvider(cert.Acme.Provider, client)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: client,
		cert:   cert,
	}, nil
}

func setProvider(p string, c *lego.Client) error {
	switch p {
	case "alidns":
		provider, err := alidns.NewDNSProvider()
		if err != nil {
			return err
		}
		return c.Challenge.SetDNS01Provider(provider)
	default:
		return errors.New("not support")
	}
}

//Create create a cert
func (c *Client) Create() (*core.Certificate, error) {
	domain := c.cert.Info.Domains[0]
	request := certificate.ObtainRequest{
		Domains: []string{"*." + domain},
		Bundle:  true,
	}
	certificates, err := c.client.Certificate.Obtain(request)
	if err != nil {
		return c.cert, err
	}
	err = c.parse(certificates.Certificate)
	if err != nil {
		return c.cert, err
	}
	c.cert.Key = base64.StdEncoding.EncodeToString(certificates.PrivateKey)
	c.cert.Cert = base64.StdEncoding.EncodeToString(certificates.Certificate)
	c.cert.CACert = base64.StdEncoding.EncodeToString(certificates.IssuerCertificate)
	return c.cert, nil
}

//Renew renew a cert
func (c *Client) Renew() (*core.Certificate, error) {
	crt, err := base64.StdEncoding.DecodeString(c.cert.Cert)
	if err != nil {
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(c.cert.Key)
	if err != nil {
		return nil, err
	}
	certificates, err := c.client.Certificate.Renew(certificate.Resource{
		Domain:      c.cert.Info.Domains[0],
		PrivateKey:  key,
		Certificate: crt,
	}, true, false, "")
	if err != nil {
		return c.cert, err
	}
	err = c.parse(certificates.Certificate)
	if err != nil {
		return c.cert, err
	}
	c.cert.Key = base64.StdEncoding.EncodeToString(certificates.PrivateKey)
	c.cert.Cert = base64.StdEncoding.EncodeToString(certificates.Certificate)
	c.cert.CACert = base64.StdEncoding.EncodeToString(certificates.IssuerCertificate)
	return c.cert, err
}

func (c *Client) parse(crt []byte) error {
	cert, err := rsa.ParseCrt(crt)
	if err != nil {
		return err
	}
	c.cert.Info = rsa.ParseCertToInfo(cert)
	return nil
}
