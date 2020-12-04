package certificate

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudflare/cfssl/cli"
	"github.com/cloudflare/cfssl/cli/genkey"
	"github.com/cloudflare/cfssl/cli/sign"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/signer"
)

//CA CA cert
type CA struct {
	Name   string
	CN     string
	Expiry string
}

// Request certificate request
type Request struct {
	Name      string
	CN        string
	O         string
	CAKey     string
	CACert    string
	Hostnames []string
}

// Certificate created key and cert data
type Certificate struct {
	Key  string
	Cert string
}

func Create(ca *CA, certReq *Request) (*Certificate, error) {
	//CA
	careq := new(csr.CertificateRequest)
	careq.KeyRequest = csr.NewKeyRequest()
	careq.KeyRequest.A = "rsa"
	careq.KeyRequest.S = 2048
	careq.CN = ca.Name
	careq.CA = &csr.CAConfig{
		Expiry: ca.Expiry,
	}
	caCert, _, caKey, err := initca.New(careq)
	if err != nil {
		return nil, err
	}
	ioutil.WriteFile(certReq.Name+"_cert", caCert, 0664)
	ioutil.WriteFile(certReq.Name+"_key", caKey, 0664)
	defer os.Remove(certReq.Name + "_cert")
	defer os.Remove(certReq.Name + "_key")
	//
	req := &csr.CertificateRequest{
		KeyRequest: csr.NewKeyRequest(),
		CN:         certReq.CN,
		Names: []csr.Name{
			{O: certReq.O},
		},
	}
	req.KeyRequest.A = "rsa"
	req.KeyRequest.S = 2048
	req.Hosts = certReq.Hostnames

	var key, csrBytes []byte
	g := &csr.Generator{Validator: genkey.Validator}
	csrBytes, key, err = g.ProcessRequest(req)
	if err != nil {
		return nil, err
	}
	config := cli.Config{
		CAFile:    certReq.Name + "_cert",
		CAKeyFile: certReq.Name + "_key",
	}
	s, err := sign.SignerFromConfig(config)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var cert []byte
	signReq := signer.SignRequest{
		Request: string(csrBytes),
		Profile: "oars-sigs",
	}

	cert, err = s.Sign(signReq)
	if err != nil {
		return nil, err
	}
	c := &Certificate{
		Key:  string(key),
		Cert: string(cert),
	}
	return c, nil
}
