package podman

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/oars-sigs/oars-cloud/core"
)

type containerCreateReq struct {
	Name           string                `json:"name,omitempty"`
	Labels         map[string]string     `json:"labels,omitempty"`
	Image          string                `json:"image,omitempty"`
	Mounts         []specs.Mount         `json:"mounts,omitempty"`
	Env            map[string]string     `json:"env,omitempty"`
	Command        core.StrSlice         `json:"command,omitempty"`
	User           string                `json:"user,omitempty"`
	ResourceLimits *specs.LinuxResources `json:"resource_limits,omitempty"`
	CapAdd         core.StrSlice         `json:"cap_add,omitempty"`
	CapDrop        core.StrSlice         `json:"cap_drop,omitempty"`
	Entrypoint     core.StrSlice         `json:"entrypoint,omitempty"`
	Devices        []specs.LinuxDevice   `json:"devices,omitempty"`
	Hostadd        []string              `json:"hostadd,omitempty"`
	Hostname       string                `json:"hostname,omitempty"`
	CniNetworks    []string              `json:"cni_networks,omitempty"`
	Privileged     bool                  `json:"privileged,omitempty"`
	StopSignal     int                   `json:"stop_signal,omitempty"`
	Sysctls        map[string]string     `json:"sysctls,omitempty"`
	WorkDir        string                `json:"work_dir,omitempty"`
}

type containerCreateResp struct {
	Id       string
	Warnings []string
}

func (c *client) Create(ctx context.Context, svc *core.ContainerService) (string, error) {
	//挂载目录
	mounts := make([]specs.Mount, 0)
	for _, v := range svc.Volumes {
		ms := strings.Split(v, ":")
		if len(ms) != 2 {
			return "", errors.New("volumes format error")
		}
		mounts = append(mounts, specs.Mount{
			Destination: ms[1],
			Source:      ms[0],
			Type:        "bind",
		})
	}
	//资源限制
	var cpu *specs.LinuxCPU
	var mem *specs.LinuxMemory
	if svc.Resources != nil {
		if svc.Resources.Memory != 0 {
			memI := svc.Resources.Memory * 1024
			mem = &specs.LinuxMemory{
				Limit: &memI,
			}
		}
		if svc.Resources.CPU != 0 {
			cpuI := int64(svc.Resources.CPU * float64(100000))
			var cpuP uint64 = 100000
			cpu = &specs.LinuxCPU{
				Quota:  &cpuI,
				Period: &cpuP,
			}
		}

	}
	env := make(map[string]string)
	//环境变量
	for _, v := range svc.Environment {
		vv := strings.SplitN(v, "=", 2)
		if len(vv) < 2 {
			return "", errors.New("env format error")
		}
		env[vv[0]] = vv[1]
	}
	req := containerCreateReq{
		Name:    svc.Name,
		Labels:  svc.Labels,
		Mounts:  mounts,
		Image:   svc.Image,
		Env:     env,
		Command: svc.Command,
		User:    svc.User,
		ResourceLimits: &specs.LinuxResources{
			Memory: mem,
			CPU:    cpu,
		},
		CapAdd:     svc.CapAdd,
		CapDrop:    svc.CapDrop,
		Entrypoint: svc.Entrypoint,
		Hostadd:    svc.ExtraHosts,
		Privileged: svc.Privileged,
		WorkDir:    svc.WorkingDir,
	}
	jsonString, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	res, err := c.Post(ctx, "/libpod/containers/create", bytes.NewBuffer(jsonString))
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(res.Body)
		return "", fmt.Errorf("unknown error, status code: %d: %s", res.StatusCode, body)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var resp containerCreateResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", err
	}
	return resp.Id, nil
}

type containerInspectResp struct {
	Names  []string
	Id     string
	Config struct {
		Labels map[string]string
	}
	NetworkSettings struct {
		IPAddress string
		Gateway   string
	}
	State struct {
		Status     string
		Running    bool
		Restarting bool
		Dead       bool
		Error      string
		ExitCode   int
		OOMKilled  bool
		Paused     bool
	}
}

//List 容器列表
func (c *client) List(ctx context.Context, all bool) ([]*core.Endpoint, error) {
	res, err := c.Get(ctx, fmt.Sprintf("/libpod/containers/json?all=%v", all))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("unknown error, status code: %d: %s", res.StatusCode, body)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var lcresp []containerCreateResp
	err = json.Unmarshal(body, &lcresp)
	if err != nil {
		return nil, err
	}
	edps := make([]*core.Endpoint, 0)
	for _, lc := range lcresp {
		res, err := c.Get(ctx, fmt.Sprintf("/libpod/containers/%s/json", lc.Id))
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusCreated {
			body, _ := ioutil.ReadAll(res.Body)
			return nil, fmt.Errorf("unknown error, status code: %d: %s", res.StatusCode, body)
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		var cn containerInspectResp
		err = json.Unmarshal(body, &cn)
		if err != nil {
			return nil, err
		}
		cname := strings.TrimPrefix(cn.Names[0], "/")
		edp := core.GetEndpointByContainerName(cname)
		edp.Labels = cn.Config.Labels
		edp.Status = &core.EndpointStatus{
			ID:          cn.Id,
			State:       cn.State.Status,
			StateDetail: cn.State.Status,
			IP:          cn.NetworkSettings.IPAddress,
			Gateway:     cn.NetworkSettings.Gateway,
		}
		edps = append(edps, edp)
	}
	return edps, nil
}

//Start 启动容器
func (c *client) Start(ctx context.Context, id string) error {
	res, err := c.Post(ctx, fmt.Sprintf("/libpod/containers/%s/start", id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("unknown error, status code: %d: %s", res.StatusCode, body)
	}
	return nil
}

//Stop 停止容器
func (c *client) Stop(ctx context.Context, id string) error {
	res, err := c.Post(ctx, fmt.Sprintf("/libpod/containers/%s/stop", id), nil)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		return nil
	}
	return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
}

//Restart 重启容器
func (c *client) Restart(ctx context.Context, id string) error {
	res, err := c.Post(ctx, fmt.Sprintf("/libpod/containers/%s/restart", id), nil)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		return nil
	}
	return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
}

//Remove 删除容器
func (c *client) Remove(ctx context.Context, id string) error {
	c.Stop(ctx, id)
	res, err := c.Delete(ctx, fmt.Sprintf("/libpod/containers/%s?force=%v", id, true))
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		return nil
	}
	return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
}

func (c *client) Log(ctx context.Context, id, tail, since string) (string, error) {
	res, err := c.Get(ctx, fmt.Sprintf("/libpod/containers/%s/logs?since=%s&tail=%s", id, since, tail))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(res.Body)
		return "", fmt.Errorf("unknown error, status code: %d: %s", res.StatusCode, body)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *client) Exec(ctx context.Context, id string, cmd string) (core.ExecResp, error) {
	sid, err := c.ExecCreate(ctx, id, ExecConfig{
		Command:      []string{cmd},
		Tty:          true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, err
	}
	return c.ExecStart(ctx, sid, ExecStartRequest{})
}
