package agent

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"

	"github.com/sirupsen/logrus"

	"github.com/oars-sigs/oars-cloud/core"
)

func (d *daemon) Create(ctx context.Context, svc *core.Service) error {
	mounts := make([]mount.Mount, 0)
	for _, v := range svc.Docker.Volumes {
		ms := strings.Split(v, ":")
		if len(ms) != 2 {
			return errors.New("volumes format error")
		}
		os.MkdirAll(ms[0], 0755)
		mounts = append(mounts, mount.Mount{
			Target: ms[1],
			Source: ms[0],
			Type:   mount.Type("bind"),
		})
	}
	for k, v := range svc.Docker.ConfigMap {
		cfgPath := d.node.WorkDir + "/configmap/" + svc.Namespace + "/" + svc.Name + "/" + strings.TrimPrefix(k, "/")
		err := os.MkdirAll(filepath.Dir(cfgPath), 0755)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(cfgPath, []byte(v), 0644)
		if err != nil {
			return err
		}
		mounts = append(mounts, mount.Mount{
			Target: k,
			Source: cfgPath,
			Type:   mount.Type("bind"),
		})
	}
	if svc.Docker.Port == nil {
		svc.Docker.Port = new(core.ContainerPort)
	}
	if svc.Docker.Port.ContainerPort == 0 {
		port, err := getFreePort()
		if err != nil {
			return err
		}
		svc.Docker.Port.ContainerPort = port
	}
	if svc.Docker.Port.Protocol == "" {
		svc.Docker.Port.Protocol = "tcp"
	}
	if svc.Docker.Environment == nil {
		svc.Docker.Environment = make([]string, 0)
	}
	port := fmt.Sprintf("SERVICE_PORT=%d", svc.Docker.Port.ContainerPort)
	svc.Docker.Environment = append(svc.Docker.Environment, port)

	if svc.Docker.Resources == nil {
		svc.Docker.Resources = new(core.ContainerResource)
	}
	cfg := &container.Config{
		Image:        svc.Docker.Image,
		AttachStdout: true,
		AttachStderr: true,
		Env:          svc.Docker.Environment,
		Cmd:          strslice.StrSlice(svc.Docker.Command),
		Shell:        strslice.StrSlice(svc.Docker.Shell),
		Labels: map[string]string{
			"serviceAddress": d.serviceAddress(svc),
			"servicePort":    fmt.Sprintf("%d", svc.Docker.Port.ContainerPort),
		},
		Entrypoint: strslice.StrSlice(svc.Docker.Entrypoint),
		StopSignal: svc.Docker.StopSignal,
		WorkingDir: svc.Docker.WorkingDir,
	}
	netMode := "host"
	if svc.Docker.NetworkMode != "" {
		netMode = svc.Docker.NetworkMode
	}

	ports := make(nat.PortMap)
	for _, p := range svc.Docker.Ports {
		portStrs := strings.Split(p, ":")
		if len(portStrs) < 2 || len(portStrs) > 3 {
			continue
		}
		hostPort := portStrs[0]
		containerPort := portStrs[1]
		hostIP := "0.0.0.0"
		if len(portStrs) == 3 {
			hostIP = portStrs[0]
			hostPort = portStrs[1]
			containerPort = portStrs[2]
		}

		ports[nat.Port(containerPort)] = []nat.PortBinding{
			{
				HostIP:   hostIP,
				HostPort: hostPort,
			},
		}
	}

	hostCfg := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: svc.Docker.Restart,
		},
		NetworkMode: container.NetworkMode(netMode),
		Mounts:      mounts,
		DNS:         []string{"127.0.0.1"},
		Resources: container.Resources{
			Memory:   svc.Docker.Resources.Memory * 1024,
			CPUQuota: int64(svc.Docker.Resources.CPU * float64(100000)),
		},
		CapAdd:       strslice.StrSlice(svc.Docker.CapAdd),
		CapDrop:      strslice.StrSlice(svc.Docker.CapDrop),
		ExtraHosts:   svc.Docker.ExtraHosts,
		Privileged:   svc.Docker.Privileged,
		SecurityOpt:  svc.Docker.SecurityOpt,
		PidMode:      container.PidMode(svc.Docker.Pid),
		Sysctls:      svc.Docker.Sysctls,
		PortBindings: ports,
	}

	hostCfg.DNS = append(hostCfg.DNS, d.node.UpDNS...)
	if svc.Docker.ImagePullPolicy == "" {
		svc.Docker.ImagePullPolicy = core.ImagePullIfNotPresent
	}
	if svc.Docker.ImagePullPolicy == core.ImagePullAlways || svc.Docker.ImagePullPolicy == core.ImagePullIfNotPresent {
		pullFlag := true
		if svc.Docker.ImagePullPolicy == core.ImagePullIfNotPresent {
			imgs, err := d.c.ImageList(ctx, types.ImageListOptions{})
			if err != nil {
				return err
			}
			imgExist := false
			for _, img := range imgs {
				for _, tag := range img.RepoTags {
					if tag == svc.Docker.Image {
						imgExist = true
					}
				}
			}
			pullFlag = !imgExist
		}
		if pullFlag {
			distributionRef, err := reference.ParseNormalizedNamed(svc.Docker.Image)
			if err != nil {
				return err
			}
			fs, err := d.c.ImagePull(ctx, distributionRef.String(), types.ImagePullOptions{})
			if err != nil {
				return err
			}
			defer fs.Close()
			_, err = d.c.ImageLoad(ctx, fs, false)
			if err != nil {
				return err
			}
		}
	}

	ct, err := d.c.ContainerCreate(ctx, cfg, hostCfg, nil, svc.Docker.Name)
	go func() {
		err = d.Start(ctx, ct.ID)
		if err != nil {
			logrus.Error(err)
		}
	}()

	return err
}

func (d *daemon) ImageList(ctx context.Context) error {
	imgs, err := d.c.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return err
	}
	for _, img := range imgs {
		fmt.Println(img.RepoTags)
	}
	return nil
}

func (d *daemon) Start(ctx context.Context, id string) error {
	return d.c.ContainerStart(ctx, id, types.ContainerStartOptions{})
}

func (d *daemon) Stop(ctx context.Context, id string) error {
	timeout := 30 * time.Second
	return d.c.ContainerStop(ctx, id, &timeout)
}

func (d *daemon) Remove(ctx context.Context, id string) error {
	d.Stop(ctx, id)
	return d.c.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true})

}

func (d *daemon) List(ctx context.Context) ([]types.Container, error) {
	return d.c.ContainerList(context.Background(), types.ContainerListOptions{All: true})
}

func (d *daemon) Restart(ctx context.Context, id string) error {
	timeout := 30 * time.Second
	return d.c.ContainerRestart(ctx, id, &timeout)
}

func (d *daemon) Log(ctx context.Context, id, tail, since string) (string, error) {
	r, err := d.c.ContainerLogs(ctx, id, types.ContainerLogsOptions{
		Tail:       tail,
		Since:      since,
		ShowStdout: true,
	})
	if err != nil {
		return "", err
	}
	defer r.Close()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

var (
	errNotFound = errors.New("No such container")
)

func (d *daemon) dockerError(err error) error {
	if strings.Contains(err.Error(), "No such container") {
		return errNotFound
	}
	return err
}
