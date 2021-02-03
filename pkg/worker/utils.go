package worker

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/oars-sigs/oars-cloud/core"

	"github.com/docker/docker/api/types"
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
		ResourceMeta: &core.ResourceMeta{
			Namespace: ns[1],
			Name:      ns[3],
		},
		Service: ns[2],
		Kind:    "container",
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

func (d *daemon) cserviceToEndpoint(cservice *core.ContainerService) *core.Endpoint {
	edp := d.getEndpointByContainerName(cservice.Name)
	status := &core.EndpointStatus{
		State: "scheduled",
		Node: core.Node{
			Hostname: d.node.Hostname,
			IP:       d.node.IP,
		},
	}
	edp.Status = status
	edp.Labels = cservice.Labels
	return edp
}

func (d *daemon) convEvent(r core.Resource, action, status, message string) *core.Event {
	event := &core.Event{
		Action:  action,
		Status:  status,
		From:    "worker-" + d.node.Hostname,
		Message: message,
		Number:  time.Now().UnixNano(),
	}
	event.GenName(r)
	return event
}

func (d *daemon) getInterface() string {
	if d.node.Interface != "" {
		return d.node.Interface
	}
	infs, _ := net.Interfaces()
	for _, inf := range infs {
		addrs, _ := inf.Addrs()
		for _, addr := range addrs {
			if addr.String() == d.node.IP {
				return inf.Name
			}
		}
	}
	return ""
}

// registryAuth returns the json marshaled, base64 encoded
// credential string that can be passed to the docker
// registry authentication header.
func registryAuth(username, password string) string {
	v := struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}{
		Username: username,
		Password: password,
	}
	buf, _ := json.Marshal(&v)
	return base64.URLEncoding.EncodeToString(buf)
}
