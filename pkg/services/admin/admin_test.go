package admin

import (
	"testing"
	"time"

	"git.cloud.gzsunrun.cn/sunrunfaas/sunrunfaas/core"
	"git.cloud.gzsunrun.cn/sunrunfaas/sunrunfaas/pkg/etcd"
)

func TestAddNamespace(t *testing.T) {
	store, err := etcd.New(&core.EtcdConfig{Endpoints: []string{"188.8.5.14:2379"}}, "sunrunfaas", 5*time.Second)
	if err != nil {
		t.Error(err)
		return
	}
	s := &service{store: store}
	reply := s.AddNamespace(map[string]string{"name": "system"})
	t.Log(*reply)
}
