package core

import (
	"encoding/json"
	"time"
)

//Certificate 证书
type Certificate struct {
	*ResourceMeta
	RootCA string           `json:"rootCA,omitempty"`
	Info   *CertInformation `json:"info,omitempty"`
	CACert string           `json:"caCert"`
	Cert   string           `json:"cert"`
	Key    string           `json:"key"`
	P12    string           `json:"p12"`
	Acme   *AcmeConfig      `json:"acme"`
}

//CertInformation cert info
type CertInformation struct {
	Country            []string  `json:"country,omitempty"`
	Organization       []string  `json:"organization,omitempty"`
	OrganizationalUnit []string  `json:"organizationalUnit,omitempty"`
	EmailAddress       []string  `json:"emailAddress,omitempty"`
	Province           []string  `json:"province,omitempty"`
	Locality           []string  `json:"locality,omitempty"`
	CommonName         string    `json:"commonName,omitempty"`
	IsCA               bool      `json:"isCA"`
	IPAddresses        []string  `json:"ipAddresses,omitempty"`
	Domains            []string  `json:"domains,omitempty"`
	Expires            int       `json:"expires,omitempty"`
	NotBefore          time.Time `json:"notBefore,omitempty"`
	NotAfter           time.Time `json:"notAfter,omitempty"`
	Password           string    `json:"password,omitempty"` //暂未实现加密证书
}

//String ...
func (r *Certificate) String() string {
	d, _ := json.Marshal(r)
	return string(d)
}

//Parse ...
func (r *Certificate) Parse(s string) error {
	return json.Unmarshal([]byte(s), r)
}

//New ...
func (r *Certificate) New() Resource {
	return &Certificate{
		ResourceMeta: new(ResourceMeta),
	}
}

//ResourceGroup ...
func (r *Certificate) ResourceGroup() string {
	return "clusters"
}

//ResourceKind ...
func (r *Certificate) ResourceKind() string {
	return "cert"
}

//ResourceKey ...
func (r *Certificate) ResourceKey() string {
	return r.Name
}

//ResourcePrefixKey ...
func (r *Certificate) ResourcePrefixKey() string {
	if r.ResourceMeta == nil {
		return ""
	}
	return r.Name
}
