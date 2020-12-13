package worker

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/oars-sigs/oars-cloud/core"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func md5V(svc *core.ContainerService) string {
	s, _ := json.Marshal(svc)
	h := md5.New()
	h.Write(s)
	return hex.EncodeToString(h.Sum(nil))
}

func (d *daemon) containerName(svc *core.Service, ed *core.ServiceEndpoint) string {
	return fmt.Sprintf("oars_%s_%s_%s", svc.Namespace, svc.Name, ed.Name)
}

func (d *daemon) containerNameByEdp(edp *core.Endpoint) string {
	return fmt.Sprintf("oars_%s_%s_%s", edp.Namespace, edp.Service, edp.Name)
}

func (d *daemon) serviceName(svc *core.Service) string {
	return fmt.Sprintf("%s.%s", svc.Namespace, svc.Name)
}

func (d *daemon) methodName(m *core.Method) string {
	return fmt.Sprintf("%s.%s.%s", m.Namespace, m.ServiceName, m.Name)
}

func (d *daemon) getEndpointByContainerName(s string) *core.Endpoint {
	ns := strings.Split(s, "_")
	return &core.Endpoint{
		Namespace: ns[1],
		Service:   ns[2],
		Name:      ns[3],
	}
}

func (d *daemon) cantainerToEndpoint(cn types.Container) *core.Endpoint {
	cname := strings.TrimPrefix(cn.Names[0], "/")
	edp := d.getEndpointByContainerName(cname)
	edp.Labels = cn.Labels
	status := &core.EndpointStatus{
		ID:          cn.ID,
		State:       cn.State,
		StateDetail: cn.Status,
		Node: core.Node{
			Hostname: d.node.Hostname,
			IP:       d.node.IP,
		},
	}
	for name, netw := range cn.NetworkSettings.Networks {
		if name == "bridge" {
			status.IP = netw.IPAddress
			status.Gateway = netw.Gateway
		} else {
			status.IP = d.node.IP
		}
	}
	edp.Status = status
	return edp
}
