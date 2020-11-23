package controller

import (
	"github.com/oars-sigs/oars-cloud/core"
)

//Start 启动controller
func Start(store core.KVStore, stopCh <-chan struct{}) {
	nodec := &nodeController{store: store}
	nodecStopCh := make(chan struct{})
	go nodec.runNodec(nodecStopCh)
	for {
		select {
		case <-stopCh:
			nodecStopCh <- struct{}{}
			break
		}
	}
}
