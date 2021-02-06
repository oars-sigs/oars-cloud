package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/registration"
)

// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func main() {
	alicfg := alidns.NewDefaultConfig()
	alicfg.APIKey = "******"
	alicfg.SecretKey = "*****"
	provider, err := alidns.NewDNSProviderConfig(alicfg)
	if err != nil {
		log.Fatal(err)
	}
	err = provider.Present("oars2.hashwing.cn", "", "dddddddddd")
	if err != nil {
		log.Fatal(err)
	}
	return
	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	myUser := MyUser{
		Email: "acme@hashwing.cn",
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	//config.CADirURL = "http://192.168.99.100:4000/directory"
	//config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		log.Fatal(err)
	}
	// We specify an HTTP port of 5002 and an TLS port of 5001 on all interfaces
	// because we aren't running as root and can't bind a listener to port 80 and 443
	// (used later when we attempt to pass challenges). Keep in mind that you still
	// need to proxy challenge traffic to port 5002 and 5001.
	// err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", "5001"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	myUser.Registration = reg

	domain := "oars2.hashwing.cn"
	request := certificate.ObtainRequest{
		Domains: []string{"*." + domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	// Each certificate comes back with the cert bytes, the bytes of the client's
	// private key, and a certificate URL. SAVE THESE TO DISK.
	//fmt.Printf("%#v\n", certificates)
	ioutil.WriteFile(domain+".key", certificates.PrivateKey, 0664)
	ioutil.WriteFile(domain+".crt", certificates.Certificate, 0664)
	ioutil.WriteFile(domain+".issuer.crt", certificates.IssuerCertificate, 0664)
	certJSON, _ := json.Marshal(certificates)
	ioutil.WriteFile(domain+".json", certJSON, 0664)
	// ... all done.
}
