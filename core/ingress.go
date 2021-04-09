package core

import (
	"encoding/json"
)

//IngressListener 入口 Listener
type IngressListener struct {
	*ResourceMeta
	Port        int              `json:"port"`
	TLSCerts    []TLSCertificate `json:"tlsCerts,omitempty"`
	DisabledTLS bool             `json:"disabledTLS"`
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

//New ...
func (l *IngressListener) New() Resource {
	return &IngressListener{
		ResourceMeta: new(ResourceMeta),
	}
}

//ResourceGroup ...
func (l *IngressListener) ResourceGroup() string {
	return "ingresses"
}

//ResourceKind ...
func (l *IngressListener) ResourceKind() string {
	return "listener"
}

//ResourceKey ...
func (l *IngressListener) ResourceKey() string {
	return l.Name
}

//ResourcePrefixKey ...
func (l *IngressListener) ResourcePrefixKey() string {
	if l.ResourceMeta == nil {
		return ""
	}
	return l.Name
}

//IngressRoute 入口路由
type IngressRoute struct {
	*ResourceMeta
	Listener string        `json:"listener"`
	Rules    []IngressRule `json:"rules"`
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

//New ...
func (route *IngressRoute) New() Resource {
	return &IngressRoute{
		ResourceMeta: new(ResourceMeta),
	}
}

//ResourceGroup ...
func (route *IngressRoute) ResourceGroup() string {
	return "ingresses"
}

//ResourceKind ...
func (route *IngressRoute) ResourceKind() string {
	return "route"
}

//ResourceKey ...
func (route *IngressRoute) ResourceKey() string {
	return "namespaces/" + route.Namespace + "/listeners/" + route.Listener + "/" + route.Name
}

//ResourcePrefixKey ...
func (route *IngressRoute) ResourcePrefixKey() string {
	if route.ResourceMeta == nil {
		return "namespaces/"
	}
	if route.Namespace != "" {
		if route.Listener != "" {
			return "namespaces/" + route.Namespace + "/listeners/" + route.Listener + "/" + route.Name
		}
		return "namespaces/" + route.Namespace + "/listeners/"
	}

	return "namespaces/"
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
