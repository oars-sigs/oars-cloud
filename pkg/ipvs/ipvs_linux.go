// +build linux

package ipvs

import (
	"net"
	"strings"
	"syscall"

	libipvs "github.com/moby/ipvs"
)

// Client ipvs client
type Client struct {
	ipvsHandle *libipvs.Handle
}

// New returns a new ipvs client.
func New() (*Client, error) {
	handle, err := libipvs.New("")
	if err != nil {
		return nil, err
	}
	return &Client{
		ipvsHandle: handle,
	}, nil
}

// AddService add a service.
func (c *Client) AddService(vs *Service) error {
	svc, err := toIPVSService(vs)
	if err != nil {
		return err
	}
	return c.ipvsHandle.NewService(svc)
}

// DeleteService ...
func (c *Client) DeleteService(vs *Service) error {
	svc, err := toIPVSService(vs)
	if err != nil {
		return err
	}
	return c.ipvsHandle.DelService(svc)
}

// GetServices ...
func (c *Client) GetServices() ([]*Service, error) {
	ipvsSvcs, err := c.ipvsHandle.GetServices()
	if err != nil {
		return nil, err
	}
	vss := make([]*Service, 0)
	for _, ipvsSvc := range ipvsSvcs {
		vs, err := toService(ipvsSvc)
		if err != nil {
			return nil, err
		}
		vss = append(vss, vs)
	}
	return vss, nil
}

// AddDestination ...
func (c *Client) AddDestination(vs *Service, rs *Destination) error {
	svc, err := toIPVSService(vs)
	if err != nil {
		return err
	}
	dst, err := toIPVSDestination(rs)
	if err != nil {
		return err
	}
	return c.ipvsHandle.NewDestination(svc, dst)
}

// DeleteDestination ...
func (c *Client) DeleteDestination(vs *Service, rs *Destination) error {
	svc, err := toIPVSService(vs)
	if err != nil {
		return err
	}
	dst, err := toIPVSDestination(rs)
	if err != nil {
		return err
	}
	return c.ipvsHandle.DelDestination(svc, dst)
}

// GetDestinations ...
func (c *Client) GetDestinations(vs *Service) ([]*Destination, error) {
	svc, err := toIPVSService(vs)
	if err != nil {
		return nil, err
	}
	dsts, err := c.ipvsHandle.GetDestinations(svc)
	if err != nil {
		return nil, err
	}
	rss := make([]*Destination, 0)
	for _, dst := range dsts {
		dst, err := toDestination(dst)
		if err != nil {
			return nil, err
		}
		rss = append(rss, dst)
	}
	return rss, nil
}

func toService(svc *libipvs.Service) (*Service, error) {
	vs := &Service{
		Address:   svc.Address.String(),
		Port:      svc.Port,
		Scheduler: svc.SchedName,
		Protocol:  protocolToString(svc.Protocol),
		Timeout:   svc.Timeout,
	}
	vs.Flags = ServiceFlags(svc.Flags &^ uint32(FlagHashed))
	return vs, nil
}

func toDestination(dst *libipvs.Destination) (*Destination, error) {
	return &Destination{
		Address: dst.Address.String(),
		Port:    dst.Port,
		Weight:  dst.Weight,
	}, nil
}

func toIPVSService(vs *Service) (*libipvs.Service, error) {
	ipvsSvc := &libipvs.Service{
		Address:       net.ParseIP(vs.Address),
		Protocol:      stringToProtocol(vs.Protocol),
		Port:          vs.Port,
		SchedName:     vs.Scheduler,
		Flags:         uint32(vs.Flags),
		Timeout:       vs.Timeout,
		AddressFamily: syscall.AF_INET,
		Netmask:       0xffffffff,
	}
	return ipvsSvc, nil
}

func toIPVSDestination(rs *Destination) (*libipvs.Destination, error) {
	return &libipvs.Destination{
		Address: net.ParseIP(rs.Address),
		Port:    rs.Port,
		Weight:  rs.Weight,
	}, nil
}

func stringToProtocol(protocol string) uint16 {
	switch strings.ToLower(protocol) {
	case "tcp":
		return uint16(syscall.IPPROTO_TCP)
	case "udp":
		return uint16(syscall.IPPROTO_UDP)
	}
	return uint16(0)
}

func protocolToString(proto uint16) string {
	switch proto {
	case syscall.IPPROTO_TCP:
		return "tcp"
	case syscall.IPPROTO_UDP:
		return "udp"
	}
	return ""
}
