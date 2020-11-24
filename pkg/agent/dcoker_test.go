package agent

import (
	"context"
	"os"
	"testing"

	"github.com/docker/docker/client"
)

func TestDocker(t *testing.T) {
	os.Setenv("DOCKER_HOST", "tcp://188.8.5.14:2375")
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Error(err)
		return
	}
	d := daemon{
		c: cli,
	}
	ctx := context.Background()
	//d.ImagePull(context.Background())
	//l, err := d.List(context.Background())
	err = d.Remove(ctx, "a97a38a0a704aa919f750735cabceca6f6bdd535f1fdf327be01cedf93c6bccb")
	t.Log(err)
	//New()
}
