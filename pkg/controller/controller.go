package controller

import (
	"github.com/oars-sigs/oars-cloud/core"
)

//Start 启动controller
func Start(store core.KVStore, cfg *core.Config, stopCh <-chan struct{}) {
	nodec := newNodec(store)
	ingressc := newIngress(store, cfg)
	nodecStopCh := make(chan struct{})
	go nodec.runNodec(nodecStopCh)
	go ingressc.run(nodecStopCh)
	for {
		select {
		case <-stopCh:
			nodecStopCh <- struct{}{}
			break
		}
	}
}
