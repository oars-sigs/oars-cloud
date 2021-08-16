package controller

import (
	"os"

	"github.com/oars-sigs/oars-cloud/core"

	log "github.com/sirupsen/logrus"
)

type worker struct {
	wfn func()
}

func (w *worker) Start() {
	w.wfn()
}
func (w *worker) Stop() {
	os.Exit(-1)
}

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
	go ingressc.run(nodecStopCh)
	log.Info("waiting a leader...")
	leader := store.LeaderController("controller")
	wfn := func() {
		log.Info("runing as a leader")
		go nodec.runNodec(nodecStopCh)
		go certc.run()
		go cronc.run(nodecStopCh)
		go svcc.run()
	}
	leader.Register(&worker{wfn})
	for {
		select {
		case <-stopCh:
			nodecStopCh <- struct{}{}
			break
		}
	}
}
