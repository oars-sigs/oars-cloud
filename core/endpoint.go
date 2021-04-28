package core

import (
	"encoding/json"
)

//Node 节点
type Node struct {
	Hostname      string `json:"hostname"`
	IP            string `json:"ip"`
	ContainerCIDR string `json:"container_cidr"`
	MAC           string `json:"mac"`
}

//Endpoint 端点
type Endpoint struct {
	*ResourceMeta
	Kind    string            `json:"kind"`
	Service string            `json:"service"`
	Labels  map[string]string `json:"labels,omitempty"`
	Status  *EndpointStatus   `json:"status,omitempty"`
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
	return &Endpoint{
		ResourceMeta: new(ResourceMeta),
	}
}

//ResourceGroup ...
func (e *Endpoint) ResourceGroup() string {
	return "services"
}

//ResourceKind ...
func (e *Endpoint) ResourceKind() string {
	return "endpoint"
}

//ResourceKey ...
func (e *Endpoint) ResourceKey() string {
	return "namespaces/" + e.Namespace + "/" + e.Service + "/" + e.Name
}

//ResourcePrefixKey ...
func (e *Endpoint) ResourcePrefixKey() string {
	if e.ResourceMeta == nil {
		return "namespaces/"
	}
	if e.Namespace != "" {
		if e.Service != "" {
			return "namespaces/" + e.Namespace + "/" + e.Service + "/" + e.Name
		}
		return "namespaces/" + e.Namespace + "/"
	}
	return "namespaces/"
}

//EndpointLogOpt 日志输出
type EndpointLogOpt struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	Tail     string `json:"tail"`
	Since    string `json:"since"`
}
