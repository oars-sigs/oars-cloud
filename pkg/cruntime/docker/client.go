package docker

import (
	"github.com/docker/docker/client"

	"github.com/oars-sigs/oars-cloud/core"
)

type daemon struct {
	client *client.Client
}

func New() (core.ContainerRuntimeInterface, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &daemon{cli}, nil
}
