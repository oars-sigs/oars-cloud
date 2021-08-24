package metrics

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/oars-sigs/oars-cloud/core"
)

// ContainerMetrics is used to track the core JSON response from the stats API
type ContainerMetrics struct {
	ID           string
	Name         string
	NetIntefaces map[string]struct {
		RxBytes   int `json:"rx_bytes"`
		RxDropped int `json:"rx_dropped"`
		RxErrors  int `json:"rx_errors"`
		RxPackets int `json:"rx_packets"`
		TxBytes   int `json:"tx_bytes"`
		TxDropped int `json:"tx_dropped"`
		TxErrors  int `json:"tx_errors"`
		TxPackets int `json:"tx_packets"`
	} `json:"networks"`
	MemoryStats struct {
		Usage int `json:"usage"`
		Limit int `json:"limit"`
		Stats struct {
			Cache int `json:"cache"`
		} `json:"stats"`
	} `json:"memory_stats"`
	CPUStats struct {
		CPUUsage struct {
			PercpuUsage       []int `json:"percpu_usage"`
			UsageInUsermode   int   `json:"usage_in_usermode"`
			TotalUsage        int   `json:"total_usage"`
			UsageInKernelmode int   `json:"usage_in_kernelmode"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
	} `json:"cpu_stats"`
	PrecpuStats struct {
		CPUUsage struct {
			PercpuUsage       []int `json:"percpu_usage"`
			UsageInUsermode   int   `json:"usage_in_usermode"`
			TotalUsage        int   `json:"total_usage"`
			UsageInKernelmode int   `json:"usage_in_kernelmode"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
}

func (e *Exporter) asyncRetrieveMetrics() ([]*core.ContainerMetrics, error) {
	// Obtain a list of running containers only
	// Docker stats API won't return stats for containers not in the running state
	edps, err := e.cri.List(context.Background(), false, true)
	if err != nil {
		logrus.Errorf("Error obtaining container listing: %v", err)
		return nil, err
	}
	containerMetrics := make([]*core.ContainerMetrics, 0)

	// range through the returned containers to obtain the statistics
	// Done due to there not yet being a '--all' option for the cli.ContainerMetrics function in the engine
	wg := new(sync.WaitGroup)
	wg.Add(len(edps))
	for _, edp := range edps {
		labels := map[string]string{
			"namespace": edp.Namespace,
			"service":   edp.Service,
			"name":      edp.Name,
		}
		go func(id string, labels map[string]string) {
			defer wg.Done()
			cm, err := e.cri.Metrics(context.Background(), id, labels)
			if err != nil {
				return
			}
			containerMetrics = append(containerMetrics, cm)

		}(edp.Status.ID, labels)
	}
	wg.Wait()
	return containerMetrics, nil
}
