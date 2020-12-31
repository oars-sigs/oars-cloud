package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
)

func TestPut(t *testing.T) {
	kv, err := New(&core.EtcdConfig{
		Endpoints: []string{"188.8.5.14:2379"},
		Prefix:    "test/",
	}, 5*time.Second)
	if err != nil {
		t.Error(err)
		return
	}
	i := 0
	for i < 10000 {
		i++
		err = kv.Put(context.Background(), core.KV{
			Key:   "test",
			Value: "dddd",
		})
		if err != nil {
			t.Error(err)
			return
		}
	}

}
