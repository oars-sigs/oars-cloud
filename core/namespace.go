package core

import (
	"context"
	"encoding/json"
)

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

//NamespaceStore namespace store
type NamespaceStore interface {
	List(ctx context.Context, arg *Namespace, opts *ListOptions) ([]*Namespace, error)
	Get(ctx context.Context, arg *Namespace, opts *GetOptions) (*Namespace, error)
	Put(ctx context.Context, arg *Namespace, opts *PutOptions) (*Namespace, error)
	Delete(ctx context.Context, arg *Namespace, opts *DeleteOptions) error
}
