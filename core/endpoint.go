package core

import (
	"context"
	"encoding/json"
)

//Node 节点
type Node struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
}

//Endpoint 端点
type Endpoint struct {
	Version   string            `json:"version"`
	Kind      string            `json:"kind"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Service   string            `json:"service"`
	Labels    map[string]string `json:"labels,omitempty"`
	Status    *EndpointStatus   `json:"status,omitempty"`
	Created   int64             `json:"created,omitempty"`
	Updated   int64             `json:"updated,omitempty"`
}

//EndpointStatus endpoint status
type EndpointStatus struct {
	State       string      `json:"state"`
	StateDetail string      `json:"stateDetail"`
	ID          string      `json:"id,omitempty"`
	IP          string      `json:"ip,omitempty"`
	Port        int         `json:"port,omitempty"`
	Gateway     string      `json:"gateway,omitempty"`
	Node        Node        `json:"node,omitempty"`
	NodeInfo    interface{} `json:"hostInfo,omitempty"`
}

//String ...
func (e *Endpoint) String() string {
	d, _ := json.Marshal(e)
	return string(d)
}

//Parse ...
func (e *Endpoint) Parse(s string) error {
	return json.Unmarshal([]byte(s), e)
}

//New ...
func (e *Endpoint) New() Resource {
	return new(Endpoint)
}

//ID ...
func (e *Endpoint) ID() string {
	return e.Namespace + "/" + e.Service + "/" + e.Name
}

//EndpointLogOpt 日志输出
type EndpointLogOpt struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	Tail     string `json:"tail"`
	Since    string `json:"since"`
}

//EndpointStore endpiont store
type EndpointStore interface {
	List(ctx context.Context, arg *Endpoint, opts *ListOptions) ([]*Endpoint, error)
	Get(ctx context.Context, arg *Endpoint, opts *GetOptions) (*Endpoint, error)
	Put(ctx context.Context, arg *Endpoint, opts *PutOptions) (*Endpoint, error)
	Delete(ctx context.Context, arg *Endpoint, opts *DeleteOptions) error
}

//EndpointLister endpiont lister
type EndpointLister interface {
	List() (ret []*Endpoint)
	Find(selector *Endpoint) (ret []*Endpoint)
}
