package controller

import (
	"github.com/oars-sigs/oars-cloud/core"
)

//Start 启动controller
func Start(store core.KVStore, cfg *core.Config, stopCh <-chan struct{}) {
	nodec := newNodec(store)
	ingressc := newIngress(store, cfg)
	certc, err := newCert(store)
	cronc := newCronc(store)
	if err != nil {
		return
	}
	svcc := newSvc(store)
	nodecStopCh := make(chan struct{})
	go nodec.runNodec(nodecStopCh)
	go ingressc.run(nodecStopCh)
	go certc.run()
	go cronc.run(nodecStopCh)
	go svcc.run()
	for {
		select {
		case <-stopCh:
			nodecStopCh <- struct{}{}
			break
		}
	}
}
