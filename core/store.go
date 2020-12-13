package core

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
