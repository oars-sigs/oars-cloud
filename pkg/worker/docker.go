package worker

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
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/utils/netutils"
)

func (d *daemon) Create(ctx context.Context, svc *core.ContainerService) (string, error) {
	edp := d.getEndpointByContainerName(svc.Name)
	mounts := make([]mount.Mount, 0)
	for _, v := range svc.Volumes {
		ms := strings.Split(v, ":")
		if len(ms) != 2 {
			return "", errors.New("volumes format error")
		}
		srcPath := ms[0]
		if !filepath.IsAbs(srcPath) {
			srcPath = fmt.Sprintf("%s/volume/%s/%s/%s", d.node.WorkDir, edp.Namespace, edp.Service, srcPath)
		}
		os.MkdirAll(srcPath, 0777)
		mounts = append(mounts, mount.Mount{
			Target: ms[1],
			Source: srcPath,
			Type:   mount.Type("bind"),
		})
	}
	for k, v := range svc.ConfigMap {
		cfgPath := d.node.WorkDir + "/configmap/" + edp.Namespace + "/" + edp.Service + "/" + strings.TrimPrefix(k, "/")
		err := os.MkdirAll(filepath.Dir(cfgPath), 0755)
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(cfgPath, []byte(v), 0644)
		if err != nil {
			return "", err
		}
		mounts = append(mounts, mount.Mount{
			Target: k,
			Source: cfgPath,
			Type:   mount.Type("bind"),
		})
	}
	if svc.Port == nil {
		svc.Port = new(core.ContainerPort)
	}
	if svc.Port.ContainerPort == 0 {
		port, err := getFreePort()
		if err != nil {
			return "", err
		}
		svc.Port.ContainerPort = port
	}
	if svc.Port.Protocol == "" {
		svc.Port.Protocol = "tcp"
	}
	if svc.Environment == nil {
		svc.Environment = make([]string, 0)
	}
	svc.Environment = append(svc.Environment, fmt.Sprintf("OARS_SERVICE_PORT=%d", svc.Port.ContainerPort))
	svc.Environment = append(svc.Environment, fmt.Sprintf("OARS_HOST_MAC=%s", d.node.MAC))
	svc.Environment = append(svc.Environment, fmt.Sprintf("OARS_HOST_IP=%s", d.node.IP))
	svc.Environment = append(svc.Environment, fmt.Sprintf("OARS_HOST_NAME=%s", d.node.Hostname))
	svc.Environment = append(svc.Environment, fmt.Sprintf("OARS_HOST_INTERFACE=%s", d.node.Interface))

	if svc.Resources == nil {
		svc.Resources = new(core.ContainerResource)
	}
	ports := make(nat.PortMap)
	portSet := make(nat.PortSet)
	for _, p := range svc.Ports {
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
		portSet[nat.Port(containerPort)] = struct{}{}
	}
	for _, expose := range svc.Expose {
		portSet[nat.Port(expose)] = struct{}{}
	}
	cfg := &container.Config{
		Image:        svc.Image,
		AttachStdout: true,
		AttachStderr: true,
		Env:          svc.Environment,
		Cmd:          strslice.StrSlice(svc.Command),
		Shell:        strslice.StrSlice(svc.Shell),
		Labels:       svc.Labels,
		Entrypoint:   strslice.StrSlice(svc.Entrypoint),
		StopSignal:   svc.StopSignal,
		WorkingDir:   svc.WorkingDir,
		ExposedPorts: portSet,
	}

	cfg.Labels[core.ServicePortLabelKey] = fmt.Sprintf("%d", svc.Port.ContainerPort)
	netMode := "bridge"
	if svc.NetworkMode != "" {
		netMode = svc.NetworkMode
	}

	hostCfg := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: svc.Restart,
		},
		NetworkMode: container.NetworkMode(netMode),
		Mounts:      mounts,
		DNS:         []string{d.node.IP},
		Resources: container.Resources{
			Memory:   svc.Resources.Memory * 1024,
			CPUQuota: int64(svc.Resources.CPU * float64(100000)),
		},
		CapAdd:       strslice.StrSlice(svc.CapAdd),
		CapDrop:      strslice.StrSlice(svc.CapDrop),
		ExtraHosts:   svc.ExtraHosts,
		Privileged:   svc.Privileged,
		SecurityOpt:  svc.SecurityOpt,
		PidMode:      container.PidMode(svc.Pid),
		Sysctls:      svc.Sysctls,
		PortBindings: ports,
	}
	if d.node.Loki.Enabled {
		labels := fmt.Sprintf("container_name={{.Name}},namespace=%s,service=%s,endpoint=%s", edp.Namespace, edp.Service, edp.Name)
		hostCfg.LogConfig = container.LogConfig{
			Type: d.node.Loki.Drive,
			Config: map[string]string{
				"loki-url":             d.node.Loki.URL,
				"max-size":             d.node.Loki.MaxSize,
				"max-file":             d.node.Loki.MaxFile,
				"loki-external-labels": labels,
			},
		}
	}
	//hostCfg.DNS = append(hostCfg.DNS, d.node.UpDNS...)
	ct, err := d.c.ContainerCreate(ctx, cfg, hostCfg, nil, svc.Name)
	if err != nil {
		return "", err
	}
	return ct.ID, err
}

func (d *daemon) ImagePull(ctx context.Context, svc *core.ContainerService) error {
	if svc.ImagePullPolicy == "" {
		svc.ImagePullPolicy = core.ImagePullIfNotPresent
	}
	if svc.ImagePullPolicy == core.ImagePullAlways || svc.ImagePullPolicy == core.ImagePullIfNotPresent {
		pullFlag := true
		if svc.ImagePullPolicy == core.ImagePullIfNotPresent {
			imgs, err := d.c.ImageList(ctx, types.ImageListOptions{})
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
			fs, err := d.c.ImagePull(ctx, distributionRef.String(), types.ImagePullOptions{RegistryAuth: svc.ImagePullAuth})
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
	return nil
}

func (d *daemon) ImageRef(ctx context.Context, image string) (string, error) {
	distributionRef, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return "", err
	}
	return distributionRef.String(), nil
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
	timeout := 10 * time.Second
	return d.c.ContainerStop(ctx, id, &timeout)
}

func (d *daemon) Remove(ctx context.Context, id string) error {
	d.Stop(ctx, id)
	return d.c.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true})

}

func (d *daemon) List(ctx context.Context) ([]types.Container, error) {
	return d.c.ContainerList(context.Background(), types.ContainerListOptions{All: true})
}

func (d *daemon) Inspect(ctx context.Context, id string) (types.ContainerJSON, error) {
	return d.c.ContainerInspect(context.Background(), id)
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
		ShowStderr: true,
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
func (d *daemon) Exec(id string, cmd string) (types.HijackedResponse, error) {
	ctx := context.Background()
	resp := types.HijackedResponse{}
	opts := types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          strings.Split(cmd, " "),
		Detach:       false,
	}
	idResp, err := d.c.ContainerExecCreate(ctx, id, opts)
	if err != nil {
		return resp, err
	}
	resp, err = d.c.ContainerExecAttach(ctx, idResp.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return resp, err
	}
	return resp, err
}

func (d *daemon) CreateNetwork(name, driver, subnet string) error {
	gateway, err := netutils.FirstSubnetIP(subnet)
	if err != nil {
		return err
	}
	nc := types.NetworkCreate{
		Driver: driver,
		IPAM: &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{
				network.IPAMConfig{
					Subnet:  subnet,
					Gateway: gateway,
				},
			},
		},
		CheckDuplicate: true,
	}
	_, err = d.c.NetworkCreate(context.Background(), name, nc)
	return err
}

func (d *daemon) ListNetworks() ([]string, error) {
	nets, err := d.c.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for _, n := range nets {
		res = append(res, n.Name)
	}
	return res, err
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
