package core

import (
	"crypto"
	"crypto/x509"

	"github.com/go-acme/lego/v4/registration"
)

//AcmeConfig acme config
type AcmeConfig struct {
	Account  *AcmeAccount      `json:"account"`
	Provider string            `json:"provider"`
	Env      map[string]string `json:"env"`
}

// AcmeAccount  a user or account type that implements acme.User
type AcmeAccount struct {
	Email        string `json:"email"`
	Registration *registration.Resource
	Key          []byte
}

// GetEmail returns email.
func (u *AcmeAccount) GetEmail() string {
	return u.Email
}

// GetRegistration returns lets encrypt registration resource.
func (u AcmeAccount) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns private key.
func (u *AcmeAccount) GetPrivateKey() crypto.PrivateKey {
	privateKey, err := x509.ParsePKCS1PrivateKey(u.Key)
	if err != nil {
		return nil
	}
	return privateKey
}
