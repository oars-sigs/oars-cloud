package metrics

import (
	"bufio"
	"context"
	"encoding/json"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
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

func (e *Exporter) asyncRetrieveMetrics() ([]*ContainerMetrics, error) {
	// Obtain a list of running containers only
	// Docker stats API won't return stats for containers not in the running state
	containers, err := e.c.ContainerList(context.Background(), types.ContainerListOptions{All: false})
	if err != nil {
		logrus.Errorf("Error obtaining container listing: %v", err)
		return nil, err
	}

	// Channels used to enable concurrent requests
	ch := make(chan *ContainerMetrics, len(containers))
	ContainerMetrics := []*ContainerMetrics{}

	// Check that there are indeed containers running we can obtain stats for
	if len(containers) == 0 {
		logrus.Errorf("No Containers returnedx from Docker socket, error: %v", err)
		return ContainerMetrics, err

	}

	// range through the returned containers to obtain the statistics
	// Done due to there not yet being a '--all' option for the cli.ContainerMetrics function in the engine
	for _, c := range containers {

		go func(cli *client.Client, id, name string) {
			retrieveContainerMetrics(*cli, id, name, ch)

		}(e.c, c.ID, c.Names[0][1:])

	}

	for {
		select {
		case r := <-ch:

			ContainerMetrics = append(ContainerMetrics, r)

			if len(ContainerMetrics) == len(containers) {
				return ContainerMetrics, nil
			}
		}

	}

}

func retrieveContainerMetrics(cli client.Client, id, name string, ch chan<- *ContainerMetrics) {

	stats, err := cli.ContainerStats(context.Background(), id, false)
	if err != nil {
		logrus.Errorf("Error obtaining container stats for %s, error: %v", id, err)
		return
	}

	s := bufio.NewScanner(stats.Body)

	for s.Scan() {

		var c *ContainerMetrics

		err := json.Unmarshal(s.Bytes(), &c)
		if err != nil {
			logrus.Errorf("Could not unmarshal the response from the docker engine for container %s. Error: %v", id, err)
			continue
		}

		// Set the container name and ID fields of the ContainerMetrics struct
		// so we can correctly report on the container when looping through later
		c.ID = id
		c.Name = name

		ch <- c
	}

	if s.Err() != nil {
		logrus.Errorf("Error handling Stats.body from Docker engine. Error: %v", s.Err())
		return
	}

}
