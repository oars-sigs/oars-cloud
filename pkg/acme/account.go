package acme

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"github.com/oars-sigs/oars-cloud/core"

	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// NewAccount creates an account.
func NewAccount(a *core.AcmeAccount) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	a.Key = x509.MarshalPKCS1PrivateKey(privateKey)
	if a.Registration == nil {
		config := lego.NewConfig(a)
		client, err := lego.NewClient(config)
		if err != nil {
			return err
		}
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return err
		}
		a.Registration = reg
	}
	return nil
}
