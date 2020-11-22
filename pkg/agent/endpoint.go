package agent

import (
	"context"
	"fmt"

	"github.com/oars-sigs/oars-cloud/core"
)

//TODO: 独立成库

func (d *daemon) putEndPoint(endpoint *core.Endpoint) error {
	ctx := context.TODO()
	v := core.KV{
		Key:   fmt.Sprintf("services/endpoint/namespaces/%s/%s/%s", endpoint.Namespace, endpoint.Service, endpoint.Hostname),
		Value: endpoint.String(),
	}
	err := d.store.Put(ctx, v)
	if err != nil {
		return err
	}
	return nil
}

func (d *daemon) deleteEndPoint(endpoint *core.Endpoint) error {
	ctx := context.TODO()
	err := d.store.Delete(ctx, fmt.Sprintf("services/endpoint/namespaces/%s/%s/%s", endpoint.Namespace, endpoint.Service, endpoint.Hostname), core.KVOption{})
	if err != nil {
		return err
	}
	return nil
}
