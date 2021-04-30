package docker

import (
	"context"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"

	"github.com/oars-sigs/oars-cloud/core"
)

func (d *daemon) ImagePull(ctx context.Context, svc *core.ContainerService) error {
	if svc.ImagePullPolicy == "" {
		svc.ImagePullPolicy = core.ImagePullIfNotPresent
	}
	if svc.ImagePullPolicy == core.ImagePullAlways || svc.ImagePullPolicy == core.ImagePullIfNotPresent {
		pullFlag := true
		if svc.ImagePullPolicy == core.ImagePullIfNotPresent {
			imgs, err := d.client.ImageList(ctx, types.ImageListOptions{})
			if err != nil {
				return err
			}
			imgExist := false
			for _, img := range imgs {
				for _, tag := range img.RepoTags {
					if tag == svc.Image {
						imgExist = true
					}
				}
			}
			pullFlag = !imgExist
		}
		if pullFlag {
			distributionRef, err := reference.ParseNormalizedNamed(svc.Image)
			if err != nil {
				return err
			}
			fs, err := d.client.ImagePull(ctx, distributionRef.String(), types.ImagePullOptions{RegistryAuth: svc.ImagePullAuth})
			if err != nil {
				return err
			}
			defer fs.Close()
			_, err = d.client.ImageLoad(ctx, fs, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
