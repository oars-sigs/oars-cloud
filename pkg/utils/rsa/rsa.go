package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	rd "math/rand"
	"net"
	"time"

	"github.com/oars-sigs/oars-cloud/core"

	"software.sslmate.com/src/go-pkcs12"
)

func init() {
	rd.Seed(time.Now().UnixNano())
}

//CreateCRT create cert
func CreateCRT(RootCa *x509.Certificate, RootKey *rsa.PrivateKey, info *core.CertInformation) (crtB []byte, keyB []byte, err error) {
	Crt := newCertificate(info)
	Key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	if RootCa == nil || RootKey == nil {
		//创建自签名证书
		crtB, err = x509.CreateCertificate(rand.Reader, Crt, Crt, &Key.PublicKey, Key)
	} else {
		//使用根证书签名
		crtB, err = x509.CreateCertificate(rand.Reader, Crt, RootCa, &Key.PublicKey, RootKey)
	}
	if err != nil {
		return
	}
	crtB = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Headers: map[string]string{}, Bytes: crtB})

	keyB = x509.MarshalPKCS1PrivateKey(Key)
	keyB = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Headers: map[string]string{}, Bytes: keyB})
	return
}

//Parse ...
func Parse(crtB, KeyB []byte) (rootcertificate *x509.Certificate, rootPrivateKey *rsa.PrivateKey, err error) {
	rootcertificate, err = ParseCrt(crtB)
	if err != nil {
		return
	}
	rootPrivateKey, err = ParseKey(KeyB)
	return
}

//ParseCrt ...
func ParseCrt(buf []byte) (*x509.Certificate, error) {
	p := &pem.Block{}
	p, buf = pem.Decode(buf)
	return x509.ParseCertificate(p.Bytes)
}

//ParseKey ...
func ParseKey(buf []byte) (*rsa.PrivateKey, error) {
	p, buf := pem.Decode(buf)
	return x509.ParsePKCS1PrivateKey(p.Bytes)
}

func newCertificate(info *core.CertInformation) *x509.Certificate {

	if len(info.Country) == 0 {
		info.Country = []string{"CN"}
	}
	if len(info.Organization) == 0 {
		info.Organization = []string{"Oars"}
	}
	if len(info.OrganizationalUnit) == 0 {
		info.OrganizationalUnit = []string{"OarsCloud"}
	}
	if len(info.Province) == 0 {
		info.Province = []string{"Guangdong"}
	}
	if len(info.Locality) == 0 {
		info.Locality = []string{"Guangzhou"}
	}
	if len(info.EmailAddress) == 0 {
		info.EmailAddress = []string{"oars@hashwing.cn"}
	}
	if info.Expires == 0 {
		info.Expires = 10
	}
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(rd.Int63()),
		Subject: pkix.Name{
			Country:            info.Country,
			Organization:       info.Organization,
			OrganizationalUnit: info.OrganizationalUnit,
			Province:           info.Province,
			CommonName:         info.CommonName,
			Locality:           info.Locality,
		},
		NotBefore:             time.Now(),                                                                 //证书的开始时间
		NotAfter:              time.Now().AddDate(info.Expires, 0, 0),                                     //证书的结束时间
		BasicConstraintsValid: true,                                                                       //基本的有效性约束
		IsCA:                  info.IsCA,                                                                  //是否是根证书
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}, //证书用途
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		EmailAddresses:        info.EmailAddress,
	}
	for _, addr := range info.IPAddresses {
		cert.IPAddresses = append(cert.IPAddresses, net.ParseIP(addr))
	}
	for _, domain := range info.Domains {
		cert.DNSNames = append(cert.DNSNames, domain)
	}
	info.NotBefore = cert.NotBefore
	info.NotAfter = cert.NotAfter
	return cert
}

//ParseCertToInfo ...
func ParseCertToInfo(cert *x509.Certificate) *core.CertInformation {
	ips := make([]string, 0)
	for _, ipAddr := range cert.IPAddresses {
		ips = append(ips, ipAddr.String())
	}
	return &core.CertInformation{
		Country:            cert.Subject.Country,
		Organization:       cert.Subject.Organization,
		OrganizationalUnit: cert.Subject.OrganizationalUnit,
		Province:           cert.Subject.Province,
		CommonName:         cert.Subject.CommonName,
		Locality:           cert.Subject.Locality,
		Expires:            cert.NotAfter.Year() - cert.NotBefore.Year(),
		NotAfter:           cert.NotAfter,
		NotBefore:          cert.NotBefore,
		IPAddresses:        ips,
		Domains:            cert.DNSNames,
		IsCA:               cert.IsCA,
	}
}

//CertToP12 ...
func CertToP12(certBuf, keyBuf []byte, certPwd string) (p12Cert string, err error) {
	caBlock, _ := pem.Decode(certBuf)
	crt, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return
	}

	keyBlock, _ := pem.Decode(keyBuf)
	priKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return
	}

	pfx, err := pkcs12.Encode(rand.Reader, priKey, crt, nil, certPwd)
	if err != nil {
		return
	}
	return base64.StdEncoding.EncodeToString(pfx), err
}
