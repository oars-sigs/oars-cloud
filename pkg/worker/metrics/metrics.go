package metrics

import (
	"fmt"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/oars-sigs/oars-cloud/core"
)

// Exporter Sets up all the runtime and metrics
type Exporter struct {
	containerMetrics map[string]*prometheus.Desc
	node             core.NodeConfig
	c                *client.Client
}

func Start(c *client.Client, node core.NodeConfig) {
	exporter := Exporter{
		containerMetrics: Return(),
		node:             node,
		c:                c,
	}
	prometheus.MustRegister(&exporter)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", node.MetricsPort), nil)
}

// Return - returns a map of metrics to be used by the exporter
func Return() map[string]*prometheus.Desc {
	labels := []string{"namespace", "service", "hostname", "name"}
	labelsInterface := append(labels, "interface")
	containerMetrics := make(map[string]*prometheus.Desc)

	// CPU Stats
	containerMetrics["cpuUsagePercent"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "cpu", "usage_percent"),
		"CPU usage percent for the specified container",
		labels, nil,
	)

	// Memory Stats
	containerMetrics["memoryUsagePercent"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "memory", "usage_percent"),
		"Current memory usage percent for the specified container",
		labels, nil,
	)
	containerMetrics["memoryUsageBytes"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "memory", "usage_bytes"),
		"Current memory usage in bytes for the specified container",
		labels, nil,
	)
	containerMetrics["memoryCacheBytes"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "memory", "cache_bytes"),
		"Current memory cache in bytes for the specified container",
		labels, nil,
	)
	containerMetrics["memoryLimit"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "memory", "limit"),
		"Memory limit as configured for the specified container",
		labels, nil,
	)

	// Network Stats
	containerMetrics["rxBytes"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "net_rx", "bytes"),
		"Network RX Bytes",
		labelsInterface, nil,
	)
	containerMetrics["rxDropped"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "net_rx", "dropped"),
		"Network RX Dropped Packets",
		labelsInterface, nil,
	)
	containerMetrics["rxErrors"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "net_rx", "errors"),
		"Network RX Packet Errors",
		labelsInterface, nil,
	)
	containerMetrics["rxPackets"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "net_rx", "packets"),
		"Network RX Packets",
		labelsInterface, nil,
	)
	containerMetrics["txBytes"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "net_tx", "bytes"),
		"Network TX Bytes",
		labelsInterface, nil,
	)
	containerMetrics["txDropped"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "net_tx", "dropped"),
		"Network TX Dropped Packets",
		labelsInterface, nil,
	)
	containerMetrics["txErrors"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "net_tx", "errors"),
		"Network TX Packet Errors",
		labelsInterface, nil,
	)
	containerMetrics["txPackets"] = prometheus.NewDesc(
		prometheus.BuildFQName("container", "net_tx", "packets"),
		"Network TX Packets",
		labelsInterface, nil,
	)
	//node
	labelsNode := []string{"hostname"}
	containerMetrics["nodeMemoryUsageBytes"] = prometheus.NewDesc(
		prometheus.BuildFQName("node", "memory", "usage_bytes"),
		"Current memory usage in bytes for the specified node",
		labelsNode, nil,
	)
	containerMetrics["nodeMemoryTotalBytes"] = prometheus.NewDesc(
		prometheus.BuildFQName("node", "memory", "total_bytes"),
		"Current memory total in bytes for the specified node",
		labelsNode, nil,
	)
	containerMetrics["nodeMemoryCacheBytes"] = prometheus.NewDesc(
		prometheus.BuildFQName("node", "memory", "cache_bytes"),
		"Current memory cache in bytes for the specified node",
		labelsNode, nil,
	)
	containerMetrics["nodeMemoryUsagePercent"] = prometheus.NewDesc(
		prometheus.BuildFQName("node", "memory", "usage_percent"),
		"memory usage percent for the specified node",
		labelsNode, nil,
	)
	containerMetrics["nodeCpuCoreNum"] = prometheus.NewDesc(
		prometheus.BuildFQName("node", "cpu", "core_num"),
		"CPU core number for the specified node",
		labelsNode, nil,
	)
	containerMetrics["nodeCpuUsagePercent"] = prometheus.NewDesc(
		prometheus.BuildFQName("node", "cpu", "usage_percent"),
		"CPU usage percent for the specified node",
		labelsNode, nil,
	)
	containerMetrics["nodeCpuLoad1"] = prometheus.NewDesc(
		prometheus.BuildFQName("node", "cpu", "load1"),
		"CPU load1 for the specified node",
		labelsNode, nil,
	)
	containerMetrics["nodeCpuLoad5"] = prometheus.NewDesc(
		prometheus.BuildFQName("node", "cpu", "load5"),
		"CPU load5 for the specified node",
		labelsNode, nil,
	)
	containerMetrics["nodeCpuLoad15"] = prometheus.NewDesc(
		prometheus.BuildFQName("node", "cpu", "load15"),
		"CPU load15 for the specified node",
		labelsNode, nil,
	)
	return containerMetrics
}
