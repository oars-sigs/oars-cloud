package core

import (
	"encoding/json"
)

type Resource interface {
	String() string
	Parse(s string) error
	New() Resource
	ID() string
}

//Namespace 命名空间
type Namespace struct {
	Version string `json:"version"`
	Name    string `json:"name" validate:"required"`
	Created int64  `json:"created,omitempty"`
	Updated int64  `json:"updated,omitempty"`
}

//String ...
func (ns *Namespace) String() string {
	d, _ := json.Marshal(ns)
	return string(d)

}

func (ns *Namespace) Parse(s string) error {
	return json.Unmarshal([]byte(s), ns)
}

//Method 方法
type Method struct {
	Version     string `json:"version"`
	Namespace   string `json:"namespace"`
	ServiceName string `json:"serviceName"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Created     int64  `json:"created,omitempty"`
	Updated     int64  `json:"updated,omitempty"`
}

//String ...
func (m *Method) String() string {
	d, _ := json.Marshal(m)
	return string(d)
}

//Parse ...
func (m *Method) Parse(s string) error {
	return json.Unmarshal([]byte(s), m)
}

//IngressListener 入口 Listener
type IngressListener struct {
	Version  string           `json:"version"`
	Name     string           `json:"name"`
	Port     int              `json:"port"`
	TLSCerts []TLSCertificate `json:"tlsCerts,omitempty"`
	Created  int64            `json:"created"`
	Updated  int64            `json:"updated"`
}

//TLSCertificate 证书
type TLSCertificate struct {
	Host string `json:"host"`
	CA   string `json:"ca,omitempty"`
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
}

//String ...
func (l *IngressListener) String() string {
	d, _ := json.Marshal(l)
	return string(d)
}

//Parse ...
func (l *IngressListener) Parse(s string) error {
	return json.Unmarshal([]byte(s), l)
}

//IngressRoute 入口路由
type IngressRoute struct {
	Version   string        `json:"version"`
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Listener  string        `json:"listener"`
	Rules     []IngressRule `json:"rules"`
	Created   int64         `json:"created"`
	Updated   int64         `json:"updated"`
}

//String ...
func (route *IngressRoute) String() string {
	d, _ := json.Marshal(route)
	return string(d)
}

//Parse ...
func (route *IngressRoute) Parse(s string) error {
	return json.Unmarshal([]byte(s), route)
}

//IngressRule Ingress规则
type IngressRule struct {
	Host string       `json:"host"`
	HTTP *IngressHTTP `json:"http,omitempty"`
	TCP  *IngressTCP  `json:"tcp,omitempty"`
}

//IngressHTTP http ingress
type IngressHTTP struct {
	Paths []IngressPath `json:"paths"`
}

//IngressTCP tcp ingress
type IngressTCP struct {
	Backend IngressBackend    `json:"backend"`
	Config  map[string]string `json:"config,omitempty"`
}

//IngressPath ingress path
type IngressPath struct {
	Path    string            `json:"path"`
	Backend IngressBackend    `json:"backend"`
	Config  map[string]string `json:"config,omitempty"`
}

//IngressBackend ingress backend
type IngressBackend struct {
	ServiceName string `json:"serviceName"`
	ServicePort int    `json:"servicePort"`
}
