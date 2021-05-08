// +build !linux

package worker

import "github.com/oars-sigs/oars-cloud/core"

func reconcileIPTables(inf string) error {
	return nil
}

func reconcileRouters(nic string, nodes []core.Node, dstRange string) error {
	return nil
}
