package rsa

import (
	"testing"

	"github.com/oars-sigs/oars-cloud/core"
)

func TestCreateCRT(t *testing.T) {
	t.Log("******************root ca***********************")
	caInfo := &core.CertInformation{
		CommonName: "OarsCloud",
		IsCA:       true,
	}
	rootCrt, rootKey, err := CreateCRT(nil, nil, caInfo)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(rootCrt))
	t.Log(string(rootKey))
	cert, err := ParseCrt(rootCrt)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("******************server***********************")
	key, err := ParseKey(rootKey)
	if err != nil {
		t.Error(err)
		return
	}
	serverInfo := &core.CertInformation{
		CommonName:  "Server",
		IPAddresses: []string{"127.0.0.1"},
		Domains:     []string{"localhost"},
	}
	serverCrt, serverKey, err := CreateCRT(cert, key, serverInfo)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(serverCrt))
	t.Log(string(serverKey))
}
