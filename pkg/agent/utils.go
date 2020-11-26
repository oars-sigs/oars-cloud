package agent

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strings"

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

func (d *daemon) containerName(svc *core.Service) string {
	return fmt.Sprintf("oars_%s_%s_%s", svc.Namespace, svc.Name, md5V(&svc.Docker))
}

func (d *daemon) serviceName(svc *core.Service) string {
	return fmt.Sprintf("%s.%s", svc.Namespace, svc.Name)
}

func (d *daemon) methodName(m *core.Method) string {
	return fmt.Sprintf("%s.%s.%s", m.Namespace, m.ServiceName, m.Name)
}

func (d *daemon) serviceAddress(svc *core.Service) string {
	return fmt.Sprintf("%s@%s:%d", svc.Docker.Port.Protocol, d.node.IP, svc.Docker.Port.ContainerPort)
}

func (d *daemon) getServiceIP(addr string) string {
	ads := strings.Split(addr, "@")
	if len(ads) < 2 {
		return addr
	}
	ps := strings.Split(ads[1], ":")
	return ps[0]
}

func (d *daemon) getEndpointByContainerName(s string) *core.Endpoint {
	ns := strings.Split(s, "_")
	return &core.Endpoint{
		Name:      s,
		Namespace: ns[1],
		Service:   ns[2],
		Hostname:  d.node.Hostname,
		HostIP:    d.node.IP,
	}
}

func (d *daemon) getCacheEndpointKey(endpoint *core.Endpoint) string {
	return endpoint.Namespace + "." + endpoint.Service + "@" + endpoint.Hostname
}
func (d *daemon) getCacheEndpointKeyBySvc(svc *core.Service) string {
	return svc.Namespace + "." + svc.Name + "@" + d.node.Hostname
}
