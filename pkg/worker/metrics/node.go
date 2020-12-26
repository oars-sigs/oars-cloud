package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

type NodeInfo struct {
	Host          *host.InfoStat `json:"host"`
	PhysicalCores int            `json:"physicalCores"`
	LogicalCores  int            `json:"logicalCores"`
	Memory        int            `json:"memory"`
}

func GetNodeInfo() (*NodeInfo, error) {
	info := &NodeInfo{}
	var err error
	info.Host, err = host.Info()
	if err != nil {
		return info, err
	}
	info.PhysicalCores, err = cpu.Counts(false)

	info.LogicalCores, err = cpu.Counts(true)

	m, err := mem.VirtualMemory()
	if err == nil {
		info.Memory = int(m.Total)
	}
	return info, err
}

func (e *Exporter) setNodeMetrics(ch chan<- prometheus.Metric) {
	m, err := mem.VirtualMemory()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["nodeMemoryUsageBytes"], prometheus.GaugeValue, float64(m.Used))
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["nodeMemoryUsagePercent"], prometheus.GaugeValue, m.UsedPercent)
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["nodeMemoryTotalBytes"], prometheus.GaugeValue, float64(m.Total))
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["nodeMemoryCacheBytes"], prometheus.GaugeValue, float64(m.Cached))
	}
	coreNum, err := cpu.Counts(true)
	if err == nil {
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["nodeCpuCoreNum"], prometheus.GaugeValue, float64(coreNum))
	}
	percents, err := cpu.Percent(time.Second, false)
	if err == nil {
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["nodeCpuUsagePercent"], prometheus.GaugeValue, percents[0])
	}
	load, err := load.Avg()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["nodeCpuLoad1"], prometheus.GaugeValue, load.Load1)
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["nodeCpuLoad5"], prometheus.GaugeValue, load.Load5)
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["nodeCpuLoad15"], prometheus.GaugeValue, load.Load15)

	}
}
