package core

import (
	"context"
	"strings"
)

//ContainerRuntimeInterface cri
type ContainerRuntimeInterface interface {
	Create(ctx context.Context, svc *ContainerService) (string, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
	Restart(ctx context.Context, id string) error
	Log(ctx context.Context, id, tail, since string) (string, error)
	List(ctx context.Context, all bool) ([]*Endpoint, error)
	Exec(ctx context.Context, id string, cmd string) (ExecResp, error)
	ImagePull(ctx context.Context, svc *ContainerService) error
	Metrics(ctx context.Context, id string, labels map[string]string) (*ContainerMetrics, error)
	CreateNetwork(ctx context.Context, name, driver, subnet string) error
	ListNetworks(ctx context.Context) ([]string, error)
}

func GetEndpointByContainerName(s string) *Endpoint {
	s = strings.TrimPrefix(s, "/")
	ns := strings.Split(s, "_")
	return &Endpoint{
		ResourceMeta: &ResourceMeta{
			Namespace: ns[1],
			Name:      ns[3],
		},
		Service: ns[2],
		Kind:    "container",
	}
}

type ExecResp interface {
	Write(p []byte) (n int, err error)
	Read(p []byte) (n int, err error)
	Close() error
}

// ContainerMetrics container metrices
type ContainerMetrics struct {
	Labels             map[string]string
	CPUUsagePercent    float64
	MemoryUsagePercent float64
	MemoryUsageBytes   int64
	MemoryCacheBytes   int64
	MemoryLimit        int64
	Network            map[string]ContainerNetworkMetric
}

//ContainerNetworkMetric container network metrices
type ContainerNetworkMetric struct {
	RxBytes   int64
	RxDropped int64
	RxErrors  int64
	RxPackets int64
	TxBytes   int64
	TxDropped int64
	TxErrors  int64
	TxPackets int64
}
