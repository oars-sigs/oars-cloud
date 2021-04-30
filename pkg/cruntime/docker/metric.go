package docker

import (
	"bufio"
	"context"
	"encoding/json"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/sirupsen/logrus"
)

func (d *daemon) Metrics(ctx context.Context, id string, labels map[string]string) (*core.ContainerMetrics, error) {
	stats, err := d.client.ContainerStats(context.Background(), id, false)
	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(stats.Body)
	for s.Scan() {
		var c *ContainerMetrics
		err := json.Unmarshal(s.Bytes(), &c)
		if err != nil {
			logrus.Errorf("Could not unmarshal the response from the docker engine for container %s. Error: %v", id, err)
			continue
		}

		cm := &core.ContainerMetrics{
			Labels:             labels,
			CPUUsagePercent:    calcCPUPercent(c),
			MemoryUsagePercent: calcMemoryPercent(c),
			MemoryUsageBytes:   c.MemoryStats.Usage,
			MemoryCacheBytes:   c.MemoryStats.Stats.Cache,
			MemoryLimit:        c.MemoryStats.Limit,
			Network:            make(map[string]core.ContainerNetworkMetric),
		}
		for name, m := range c.NetIntefaces {
			cm.Network[name] = core.ContainerNetworkMetric{
				RxBytes:   m.RxBytes,
				RxDropped: m.RxDropped,
				RxErrors:  m.RxErrors,
				RxPackets: m.RxPackets,
				TxBytes:   m.TxBytes,
				TxDropped: m.TxDropped,
				TxErrors:  m.TxErrors,
				TxPackets: m.TxPackets,
			}
		}
		return cm, nil
	}

	return nil, s.Err()

}

// ContainerMetrics is used to track the core JSON response from the stats API
type ContainerMetrics struct {
	ID           string
	Name         string
	NetIntefaces map[string]struct {
		RxBytes   int64 `json:"rx_bytes"`
		RxDropped int64 `json:"rx_dropped"`
		RxErrors  int64 `json:"rx_errors"`
		RxPackets int64 `json:"rx_packets"`
		TxBytes   int64 `json:"tx_bytes"`
		TxDropped int64 `json:"tx_dropped"`
		TxErrors  int64 `json:"tx_errors"`
		TxPackets int64 `json:"tx_packets"`
	} `json:"networks"`
	MemoryStats struct {
		Usage int64 `json:"usage"`
		Limit int64 `json:"limit"`
		Stats struct {
			Cache int64 `json:"cache"`
		} `json:"stats"`
	} `json:"memory_stats"`
	CPUStats struct {
		CPUUsage struct {
			PercpuUsage       []int64 `json:"percpu_usage"`
			UsageInUsermode   int64   `json:"usage_in_usermode"`
			TotalUsage        int64   `json:"total_usage"`
			UsageInKernelmode int64   `json:"usage_in_kernelmode"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
	} `json:"cpu_stats"`
	PrecpuStats struct {
		CPUUsage struct {
			PercpuUsage       []int64 `json:"percpu_usage"`
			UsageInUsermode   int64   `json:"usage_in_usermode"`
			TotalUsage        int64   `json:"total_usage"`
			UsageInKernelmode int64   `json:"usage_in_kernelmode"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
}

func calcCPUPercent(stats *ContainerMetrics) float64 {

	var CPUPercent float64

	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PrecpuStats.CPUUsage.TotalUsage)
	sysDelta := float64(stats.CPUStats.SystemCPUUsage - stats.PrecpuStats.SystemCPUUsage)

	if sysDelta > 0.0 && cpuDelta > 0.0 {
		CPUPercent = (cpuDelta / sysDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	return CPUPercent
}

func calcMemoryPercent(stats *ContainerMetrics) float64 {
	return float64(stats.MemoryStats.Usage) * 100.0 / float64(stats.MemoryStats.Limit)
}
