package metrics

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
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
