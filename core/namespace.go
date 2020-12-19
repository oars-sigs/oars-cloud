package core

import (
	"encoding/json"
)

//Namespace 命名空间
type Namespace struct {
	*ResourceMeta
}

//String ...
func (ns *Namespace) String() string {
	d, _ := json.Marshal(ns)
	return string(d)
}

func (ns *Namespace) Parse(s string) error {
	return json.Unmarshal([]byte(s), ns)
}

//New ...
func (ns *Namespace) New() Resource {
	return &Namespace{
		ResourceMeta: new(ResourceMeta),
	}
}

//ResourceGroup ...
func (ns *Namespace) ResourceGroup() string {
	return "clusters"
}

//ResourceKind ...
func (ns *Namespace) ResourceKind() string {
	return "namespace"
}

//ResourceKey ...
func (ns *Namespace) ResourceKey() string {
	return ns.Name
}

//ResourcePrefixKey ...
func (ns *Namespace) ResourcePrefixKey() string {
	if ns.ResourceMeta == nil {
		return ""
	}
	return ns.Name
}
