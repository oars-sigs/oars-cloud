package core

import (
	"context"
)

type ResourceMeta struct {
	Version   string `json:"version,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
	Created   int64  `json:"created,omitempty"`
	Updated   int64  `json:"updated,omitempty"`
}

//SetCreated ...
func (m *ResourceMeta) SetCreated(t int64) {
	m.Created = t
}

//SetUpdated ...
func (m *ResourceMeta) SetUpdated(t int64) {
	m.Updated = t
}

//GetCreated ...
func (m *ResourceMeta) GetCreated() int64 {
	return m.Created
}

//GetUpdated ...
func (m *ResourceMeta) GetUpdated() int64 {
	return m.Updated
}

type Resource interface {
	String() string
	Parse(s string) error
	New() Resource
	ResourceKind() string
	ResourceGroup() string
	ResourceKey() string
	ResourcePrefixKey() string
	SetCreated(t int64)
	SetUpdated(t int64)
	GetCreated() int64
	GetUpdated() int64
}

//ResourceStore resource store
type ResourceStore interface {
	List(ctx context.Context, arg Resource, opts *ListOptions) ([]Resource, error)
	Get(ctx context.Context, arg Resource, opts *GetOptions) (Resource, error)
	Put(ctx context.Context, arg Resource, opts *PutOptions) (Resource, error)
	Delete(ctx context.Context, arg Resource, opts *DeleteOptions) error
}

type ListOptions struct{}
type GetOptions struct{}
type DeleteOptions struct{}
type PutOptions struct{}
type CreateOptions struct{}
type UpdateOptions struct{}

type ResourceLister interface {
	List() ([]Resource, bool)
}

type ResourceEventHandle struct {
	Trigger     chan struct{}
	Interceptor func(put bool, current, pre Resource) (Resource, bool, error)
}
