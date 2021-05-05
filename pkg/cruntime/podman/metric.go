package podman

import (
	"context"

	"github.com/oars-sigs/oars-cloud/core"
)

func (c *client) Metrics(ctx context.Context, id string, labels map[string]string) (*core.ContainerMetrics, error) {
	return nil, nil
}
