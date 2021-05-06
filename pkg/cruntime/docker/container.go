package docker

import (
	"bufio"
	"context"
	"errors"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"

	"github.com/oars-sigs/oars-cloud/core"
)

//Create 创建容器
func (d *daemon) Create(ctx context.Context, svc *core.ContainerService) (string, error) {
	//挂载目录
	mounts := make([]mount.Mount, 0)
	for _, v := range svc.Volumes {
		ms := strings.Split(v, ":")
		if len(ms) != 2 {
			return "", errors.New("volumes format error")
		}
		mounts = append(mounts, mount.Mount{
			Target: ms[1],
			Source: ms[0],
			Type:   mount.Type("bind"),
		})
	}

	//资源限制
	if svc.Resources == nil {
		svc.Resources = new(core.ContainerResource)
	}
	//绑定端口
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
	//容器配置
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

	netMode := "bridge"
	if svc.NetworkMode != "" {
		netMode = svc.NetworkMode
	}

	//主机配置
	hostCfg := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: svc.Restart,
		},
		NetworkMode: container.NetworkMode(netMode),
		Mounts:      mounts,
		DNS:         svc.DNS,
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
	//创建容器
	ct, err := d.client.ContainerCreate(ctx, cfg, hostCfg, nil, svc.Name)
	if err != nil {
		return "", err
	}
	return ct.ID, err
}

//Start 启动容器
func (d *daemon) Start(ctx context.Context, id string) error {
	return d.client.ContainerStart(ctx, id, types.ContainerStartOptions{})
}

//Stop 停止容器
func (d *daemon) Stop(ctx context.Context, id string) error {
	timeout := 10 * time.Second
	return d.client.ContainerStop(ctx, id, &timeout)
}

//Remove 删除容器
func (d *daemon) Remove(ctx context.Context, id string) error {
	d.Stop(ctx, id)
	return d.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true})
}

//Restart 重启容器
func (d *daemon) Restart(ctx context.Context, id string) error {
	timeout := 30 * time.Second
	return d.client.ContainerRestart(ctx, id, &timeout)
}

func (d *daemon) Log(ctx context.Context, id, tail, since string) (string, error) {
	r, err := d.client.ContainerLogs(ctx, id, types.ContainerLogsOptions{
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

//List 容器列表
func (d *daemon) List(ctx context.Context, all bool) ([]*core.Endpoint, error) {
	cs, err := d.client.ContainerList(context.Background(), types.ContainerListOptions{All: all})
	if err != nil {
		return nil, err
	}
	edps := make([]*core.Endpoint, 0)
	for _, cn := range cs {
		_, ok := cn.Labels[core.CreatorLabelKey]
		if !ok {
			continue
		}
		cname := strings.TrimPrefix(cn.Names[0], "/")
		edp := core.GetEndpointByContainerName(cname)
		edp.Labels = cn.Labels
		status := &core.EndpointStatus{
			ID:          cn.ID,
			State:       cn.State,
			StateDetail: cn.Status,
		}
		for name, netw := range cn.NetworkSettings.Networks {
			status.Network = name
			if name != "host" {
				status.IP = netw.IPAddress
				status.Gateway = netw.Gateway
			}

		}
		edp.Status = status
		edps = append(edps, edp)
	}
	return edps, nil
}

// HijackedResponse holds connection information for a hijacked request.
type HijackedResponse struct {
	Conn   net.Conn
	Reader *bufio.Reader
}

// Close closes the hijacked connection and reader.
func (h *HijackedResponse) Close() error {
	return h.Conn.Close()
}

func (h *HijackedResponse) Write(p []byte) (n int, err error) {
	return h.Conn.Write(p)
}

func (h *HijackedResponse) Read(p []byte) (n int, err error) {
	return h.Reader.Read(p)
}

func (d *daemon) Exec(ctx context.Context, id string, cmd string) (core.ExecResp, error) {
	opts := types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          strings.Split(cmd, " "),
		Detach:       false,
	}
	idResp, err := d.client.ContainerExecCreate(ctx, id, opts)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.ContainerExecAttach(ctx, idResp.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return nil, err
	}
	return &HijackedResponse{
		Conn:   resp.Conn,
		Reader: resp.Reader,
	}, err
}
