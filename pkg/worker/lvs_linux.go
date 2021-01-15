// +build linux

package worker

import (
	"strconv"
	"strings"
	"time"

	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/e"
	"github.com/oars-sigs/oars-cloud/pkg/ipvs"
	"github.com/sirupsen/logrus"

	"github.com/vishvananda/netlink"
)

type lvs struct {
	ipvsLink   netlink.Link
	ipvsClient *ipvs.Client
	svcLister  core.ResourceLister
	edpLister  core.ResourceLister
}

func startLVS(svcLister, edpLister core.ResourceLister) error {
	links, err := netlink.LinkList()
	if err != nil {
		return err
	}
	isExist := false
	for _, link := range links {
		if link.Attrs().Name == core.IPVSNicName {
			isExist = true
			break
		}
	}
	if !isExist {
		la := netlink.NewLinkAttrs()
		la.Name = core.IPVSNicName
		err := netlink.LinkAdd(&netlink.Dummy{LinkAttrs: la})
		if err != nil {
			return err
		}
	}
	ipvsLink, err := netlink.LinkByName(core.IPVSNicName)
	if err != nil {
		return err
	}
	ipvsClient, err := ipvs.New()
	if err != nil {
		return err
	}
	l := &lvs{
		ipvsLink:   ipvsLink,
		svcLister:  svcLister,
		edpLister:  edpLister,
		ipvsClient: ipvsClient,
	}
	go l.start()
	return nil

}

func (l *lvs) start() {
	t := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-t.C:
			err := l.syncService()
			if err != nil {
				logrus.Error(err)
			}
		}
	}
}

func (l *lvs) syncService() error {
	svcRess, ok := l.svcLister.List()
	if !ok {
		return nil
	}
	edpRess, ok := l.edpLister.List()
	if !ok {
		return nil
	}
	ipvsSvcs, err := l.ipvsClient.GetServices()
	if err != nil {
		return err
	}
	addrs, err := netlink.AddrList(l.ipvsLink, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	//add addr to oars-ipvs interface and lvs server
	for _, res := range svcRess {
		svc := res.(*core.Service)
		if svc.VirtualServer == nil {
			continue
		}
		addrExist := false
		for _, addr := range addrs {
			if addr.IP.String() == svc.VirtualServer.ClusterIP {
				addrExist = true
			}
		}
		if !addrExist {
			err = l.addAddr(svc.VirtualServer.ClusterIP + "/32")
			if err != nil {
				logrus.Error(err)
				continue
			}
		}

		dstIPs := make([]string, 0)
		for _, res := range edpRess {
			edp := res.(*core.Endpoint)
			if edp.Service == svc.Name && edp.Status.State == "running" {
				dstIPs = append(dstIPs, edp.Status.IP)
			}
		}
		l.addService(svc.VirtualServer, dstIPs, ipvsSvcs)
	}

	//gc lvs servers
	for _, ipvsSvc := range ipvsSvcs {
		vsExist := false
		for _, res := range svcRess {
			svc := res.(*core.Service)
			if svc.VirtualServer == nil {
				continue
			}
			if ipvsSvc.Address == svc.VirtualServer.ClusterIP {
				vsExist = true
				break
			}
		}
		if !vsExist {
			err = l.ipvsClient.DeleteService(ipvsSvc)
			if err != nil {
				logrus.Error(err)
				continue
			}
		}
	}

	//gc addrs
	for _, addr := range addrs {
		svcExist := false
		for _, res := range svcRess {
			svc := res.(*core.Service)
			if svc.VirtualServer == nil {
				continue
			}
			if addr.IP.String() == svc.VirtualServer.ClusterIP {
				svcExist = true
			}
		}
		if !svcExist {
			err = netlink.AddrDel(l.ipvsLink, &addr)
			if err != nil {
				logrus.Error(err)
				continue
			}
		}

	}
	return nil
}

func (l *lvs) addAddr(ip string) error {
	clusterAddr, err := netlink.ParseAddr(ip)
	if err != nil {
		return err
	}
	return netlink.AddrAdd(l.ipvsLink, clusterAddr)
}

func (l *lvs) addService(vs *core.VirtualServer, dstIPs []string, ipvsSvcs []*ipvs.Service) error {
	for _, portStr := range vs.Ports {
		protocol, svcPort, targetPort, err := parsePort(portStr)
		if err != nil {
			logrus.Error(err)
			continue
		}
		ipvsSvc := &ipvs.Service{
			Address:   vs.ClusterIP,
			Protocol:  protocol,
			Port:      uint16(svcPort),
			Scheduler: "rr",
			Flags:     0,
			Timeout:   0,
		}
		vsExist := false
		for _, oipvsSvc := range ipvsSvcs {
			if oipvsSvc.Address == vs.ClusterIP && int(oipvsSvc.Port) == svcPort {
				vsExist = true
				ipvsSvc = oipvsSvc
				break
			}
		}
		if !vsExist {
			err = l.ipvsClient.AddService(ipvsSvc)
			if err != nil {
				logrus.Error(err)
				continue
			}
		}

		ipvsDsts, err := l.ipvsClient.GetDestinations(ipvsSvc)
		if err != nil {
			logrus.Error(err)
		}
		for _, dstIP := range dstIPs {
			isExist := false
			for _, dst := range ipvsDsts {
				if dst.Address == dstIP && targetPort == int(dst.Port) {
					isExist = true
				}
			}
			if !isExist {
				ipvsDst := &ipvs.Destination{
					Address: dstIP,
					Port:    uint16(targetPort),
					Weight:  1,
				}
				err = l.ipvsClient.AddDestination(ipvsSvc, ipvsDst)
				if err != nil {
					logrus.Error(err)
				}
			}
		}
		for _, dst := range ipvsDsts {
			isExist := false
			for _, dstIP := range dstIPs {
				if dst.Address == dstIP && targetPort == int(dst.Port) {
					isExist = true
				}
			}
			if !isExist {
				err = l.ipvsClient.DeleteDestination(ipvsSvc, dst)
				if err != nil {
					logrus.Error(err)
				}
			}
		}
	}

	return nil
}

func parsePort(portStr string) (string, int, int, error) {
	ports := strings.Split(portStr, ":")
	protocol := "tcp"
	targetPort := 0
	svcPort, err := strconv.Atoi(ports[0])
	if err != nil {
		return protocol, svcPort, targetPort, e.ErrInvalidPortFormat
	}
	switch len(ports) {
	case 1:
		targetPort, err = strconv.Atoi(ports[0])
		if err != nil {
			return protocol, svcPort, targetPort, e.ErrInvalidPortFormat
		}
	case 2:
		targetPort, err = strconv.Atoi(ports[1])
		if err != nil {
			return protocol, svcPort, targetPort, e.ErrInvalidPortFormat
		}
	case 3:
		targetPort, err = strconv.Atoi(ports[1])
		if err != nil {
			return protocol, svcPort, targetPort, e.ErrInvalidPortFormat
		}
		protocol = ports[2]
	default:
		return protocol, svcPort, targetPort, e.ErrInvalidPortFormat
	}
	return protocol, svcPort, targetPort, nil
}
