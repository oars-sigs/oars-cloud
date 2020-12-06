package core

import (
	"bytes"
	"encoding/json"
	"html/template"
)

//Namespace 命名空间
type Namespace struct {
	Version string `json:"version"`
	Name    string `json:"name"`
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

//Service 服务
type Service struct {
	Version   string           `json:"version"`
	Namespace string           `json:"namespace"`
	Name      string           `json:"name"`
	Kind      string           `json:"kind"`
	Endpoints []Endpoint       `json:"endpoints"`
	Docker    ContainerService `json:"docker,omitempty"`
	RuntimeID string           `json:"-"`
	Created   int64            `json:"created,omitempty"`
	Updated   int64            `json:"updated,omitempty"`
}

//String ...
func (svc *Service) String() string {
	d, _ := json.Marshal(svc)
	return string(d)
}

func (svc *Service) Parse(s string) error {
	return json.Unmarshal([]byte(s), svc)
}

func (svc *Service) ParseTpl(endpoint Endpoint, values map[string]interface{}) error {
	if svc.Kind == "docker" {
		tmpl := template.New("tpl")
		tmpl, err := tmpl.Parse(svc.Docker.String())
		if err != nil {
			return err
		}
		vars := ContainerValues{
			Global:   values,
			Endpoint: endpoint,
		}
		var b bytes.Buffer
		err = tmpl.Execute(&b, vars)
		if err != nil {
			return err
		}
		csvc := new(ContainerService)
		err = csvc.Parse(b.String())
		if err != nil {
			return err
		}
		svc.Docker = *csvc
	}
	return nil
}

//Endpoint 端点
type Endpoint struct {
	Version   string                 `json:"version"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Namespace string                 `json:"namespace,omitempty"`
	Service   string                 `json:"service,omitempty"`
	Status    string                 `json:"status,omitempty"`
	State     string                 `json:"state,omitempty"`
	Port      int                    `json:"port,omitempty"`
	Hostname  string                 `json:"hostname,omitempty"`
	HostIP    string                 `json:"hostIP,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
	NodeInfo  interface{}            `json:"hostInfo,omitempty"`
	Created   int64                  `json:"created,omitempty"`
	Updated   int64                  `json:"updated,omitempty"`
}

//String ...
func (e *Endpoint) String() string {
	d, _ := json.Marshal(e)
	return string(d)
}

func (e *Endpoint) Parse(s string) error {
	return json.Unmarshal([]byte(s), e)
}

//EndpointLogOpt 日志输出
type EndpointLogOpt struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	Tail     string `json:"tail"`
	Since    string `json:"since"`
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
