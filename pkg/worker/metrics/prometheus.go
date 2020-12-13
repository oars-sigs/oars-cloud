package metrics

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Describe - loops through the API metrics and passes them to prometheus.Describe
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

	for _, m := range e.containerMetrics {
		ch <- m
	}

}

// Collect function, called on by Prometheus Client library
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	logrus.Info("Metric collection requested")

	metrics, err := e.asyncRetrieveMetrics()

	if err != nil {
		logrus.Error("Errors in collection")
	}

	if len(metrics) == 0 {
		logrus.Info("No valid container metrics to process")
		return
	}

	for _, b := range metrics {
		e.setPrometheusMetrics(b, ch)
	}

	logrus.Info("Metric collection completed")

}

// setPrometheusMetrics takes the pointer to ContainerMetrics and uses the data to set the guages and counters
func (e *Exporter) setPrometheusMetrics(stats *ContainerMetrics, ch chan<- prometheus.Metric) {
	if !strings.HasPrefix(stats.Name, "sunrunfaas_") {
		return
	}
	names := strings.Split(stats.Name, "_")
	if len(names) != 4 {
		return
	}
	labels := []string{names[1], names[2], e.node.Hostname, stats.ID, stats.Name}

	// Set CPU metrics
	ch <- prometheus.MustNewConstMetric(e.containerMetrics["cpuUsagePercent"], prometheus.GaugeValue, calcCPUPercent(stats), labels...)

	// Set Memory metrics
	ch <- prometheus.MustNewConstMetric(e.containerMetrics["memoryUsagePercent"], prometheus.GaugeValue, calcMemoryPercent(stats), labels...)
	ch <- prometheus.MustNewConstMetric(e.containerMetrics["memoryUsageBytes"], prometheus.GaugeValue, float64(stats.MemoryStats.Usage), labels...)
	ch <- prometheus.MustNewConstMetric(e.containerMetrics["memoryCacheBytes"], prometheus.GaugeValue, float64(stats.MemoryStats.Stats.Cache), labels...)
	ch <- prometheus.MustNewConstMetric(e.containerMetrics["memoryLimit"], prometheus.GaugeValue, float64(stats.MemoryStats.Limit), labels...)

	if len(stats.NetIntefaces) == 0 {
		logrus.Infof("No network interfaces detected for container %s", stats.Name)
	}

	// Network interface stats (loop through the map of returned interfaces)
	for key, net := range stats.NetIntefaces {
		labelsInterface := append(labels, key)
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["rxBytes"], prometheus.GaugeValue, float64(net.RxBytes), labelsInterface...)
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["rxDropped"], prometheus.GaugeValue, float64(net.RxDropped), labelsInterface...)
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["rxErrors"], prometheus.GaugeValue, float64(net.RxErrors), labelsInterface...)
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["rxPackets"], prometheus.GaugeValue, float64(net.RxPackets), labelsInterface...)
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["txBytes"], prometheus.GaugeValue, float64(net.TxBytes), labelsInterface...)
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["txDropped"], prometheus.GaugeValue, float64(net.TxDropped), labelsInterface...)
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["txErrors"], prometheus.GaugeValue, float64(net.TxErrors), labelsInterface...)
		ch <- prometheus.MustNewConstMetric(e.containerMetrics["txPackets"], prometheus.GaugeValue, float64(net.TxPackets), labelsInterface...)
	}

}
