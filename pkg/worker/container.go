package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/oars-sigs/oars-cloud/core"
)

func (d *daemon) Create(ctx context.Context, svc *core.ContainerService) (string, error) {
	edp := core.GetEndpointByContainerName(svc.Name)
	volumes := make([]string, 0)
	for _, v := range svc.Volumes {
		ms := strings.Split(v, ":")
		if len(ms) != 2 {
			return "", errors.New("volumes format error")
		}
		if !filepath.IsAbs(ms[0]) {
			ms[0] = fmt.Sprintf("%s/volume/%s/%s/%s", d.node.WorkDir, edp.Namespace, edp.Service, ms[0])
		}
		os.MkdirAll(ms[0], 0755)
		volumes = append(volumes, strings.Join(ms, ":"))
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
		volumes = append(volumes, cfgPath+":"+k)
	}
	svc.Volumes = volumes
	if svc.Environment == nil {
		svc.Environment = make([]string, 0)
	}
	svc.Environment = append(svc.Environment, fmt.Sprintf("OARS_HOST_MAC=%s", d.node.MAC))
	svc.Environment = append(svc.Environment, fmt.Sprintf("OARS_HOST_IP=%s", d.node.IP))
	svc.Environment = append(svc.Environment, fmt.Sprintf("OARS_HOST_NAME=%s", d.node.Hostname))
	svc.Environment = append(svc.Environment, fmt.Sprintf("OARS_HOST_INTERFACE=%s", d.node.Interface))

	if svc.Resources == nil {
		svc.Resources = new(core.ContainerResource)
	}
	if svc.NetworkMode == "" {
		svc.NetworkMode = "bridge"
	}
	svc.DNS = []string{d.node.IP}
	svc.DNSSearch = []string{edp.Namespace}
	parentContainer := ""
	if strings.HasPrefix(svc.NetworkMode, "service:") {
		parentContainer = d.gencontainerName(edp.Namespace, strings.TrimPrefix(svc.NetworkMode, "service:"), edp.Name)
		svc.NetworkMode = fmt.Sprintf("container:%s", parentContainer)
	}
	if strings.HasPrefix(svc.NetworkMode, "container:") {
		svc.DNS = []string{}
		svc.DNSSearch = []string{}
	}
	//svc.DNS = append(svc.DNS, d.node.UpDNS...)
	cid, err := d.cri.Create(ctx, svc)
	if err != nil {
		return "", err
	}
	//set depend
	if parentContainer != "" {
		d.setDepend(parentContainer, cid)
	}

	return cid, err
}

func (d *daemon) ImagePull(ctx context.Context, svc *core.ContainerService) error {
	return d.cri.ImagePull(ctx, svc)
}

func (d *daemon) Start(ctx context.Context, id string) error {
	return d.cri.Start(ctx, id)
}

func (d *daemon) Stop(ctx context.Context, id string) error {
	return d.cri.Stop(ctx, id)
}

func (d *daemon) Remove(ctx context.Context, id string) error {
	return d.cri.Remove(ctx, id)

}

func (d *daemon) List(ctx context.Context) ([]*core.Endpoint, error) {
	return d.cri.List(ctx, true)
}

func (d *daemon) Restart(ctx context.Context, id string) error {
	return d.cri.Restart(ctx, id)
}

func (d *daemon) Log(ctx context.Context, id, tail, since string) (string, error) {
	return d.cri.Log(ctx, id, tail, since)
}
func (d *daemon) Exec(id string, cmd string, w, h uint) (core.ExecResp, error) {
	ctx := context.Background()
	return d.cri.Exec(ctx, id, cmd, w, h)
}

func (d *daemon) CreateNetwork(name, driver, subnet string) error {
	return d.cri.CreateNetwork(context.Background(), name, driver, subnet)
}

func (d *daemon) ListNetworks() ([]string, error) {
	return d.cri.ListNetworks(context.Background())

}

func (d *daemon) LogStream(id, tail, since string) (io.ReadCloser, error) {
	return d.cri.LogStream(context.Background(), id, tail, since)
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
