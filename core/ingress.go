package core

import (
	"context"
	"encoding/json"
)

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

//New ...
func (l *IngressListener) New() Resource {
	return new(IngressListener)
}

//ID ...
func (l *IngressListener) ID() string {
	return l.Name
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

//New ...
func (route *IngressRoute) New() Resource {
	return new(IngressRoute)
}

//ID ...
func (route *IngressRoute) ID() string {
	return route.Namespace + "/" + route.Name
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

//IngressStore ingress listener store
type IngressListenerStore interface {
	List(ctx context.Context, arg *IngressListener, opts *ListOptions) ([]*IngressListener, error)
	Get(ctx context.Context, arg *IngressListener, opts *GetOptions) (*IngressListener, error)
	Put(ctx context.Context, arg *IngressListener, opts *PutOptions) (*IngressListener, error)
	Delete(ctx context.Context, arg *IngressListener, opts *DeleteOptions) error
}

//IngressListenerLister ingress listener lister
type IngressListenerLister interface {
	List() (ret []*IngressListener)
	Find(selector *IngressListener) (ret []*IngressListener)
}

//IngressRouteStore ingress route store
type IngressRouteStore interface {
	List(ctx context.Context, arg *IngressRoute, opts *ListOptions) ([]*IngressRoute, error)
	Get(ctx context.Context, arg *IngressRoute, opts *GetOptions) (*IngressRoute, error)
	Put(ctx context.Context, arg *IngressRoute, opts *PutOptions) (*IngressRoute, error)
	Delete(ctx context.Context, arg *IngressRoute, opts *DeleteOptions) error
}

//IngressRouteLister ingress route lister
type IngressRouteLister interface {
	List() (ret []*IngressRoute)
	Find(selector *IngressRoute) (ret []*IngressRoute)
}
