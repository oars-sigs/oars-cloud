// +build !linux

package worker

import (
	"github.com/oars-sigs/oars-cloud/core"
)

func startLVS(svcLister, edpLister core.ResourceLister) error {
	return nil
}

func reconcileRouters(nic string, cidrs []string) error {
	return nil
}
