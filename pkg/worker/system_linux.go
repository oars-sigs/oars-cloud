// +build linux

package worker

import (
	"net"

	"github.com/coreos/go-iptables/iptables"
	"github.com/oars-sigs/oars-cloud/core"
	"github.com/oars-sigs/oars-cloud/pkg/utils/netutils"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func reconcileIPTables(inf string) error {
	// Init iptable
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	exists, err := ipt.Exists("filter", "FORWARD", "-i", inf, "-j", "ACCEPT")
	if err != nil {
		return err
	}
	if !exists {
		if err := ipt.Insert("filter", "FORWARD", 1, "-i", inf, "-j", "ACCEPT"); err != nil {
			return err
		}
	}
	return nil
}

func reconcileRouters(link string, nodes []core.Node, dstRange string) (err error) {
	nic, err := netlink.LinkByName(link)
	if err != nil {
		return
	}
	existRoutes, err := netlink.RouteList(nic, netlink.FAMILY_V4)
	if err != nil {
		return
	}
	toDel := make([]string, 0)
	toAdd := make([]core.Node, 0)

	for _, route := range existRoutes {
		if route.Dst == nil {
			continue
		}
		if route.Scope == netlink.SCOPE_LINK {
			continue
		}

		found := false
		for _, node := range nodes {
			if route.Dst.String() == node.ContainerCIDR {
				found = true
				break
			}
		}
		if !found {
			toDel = append(toDel, route.Dst.String())
		}
	}

	for _, node := range nodes {
		found := false
		for _, r := range existRoutes {
			if r.Dst == nil {
				continue
			}
			if r.Dst.String() == node.ContainerCIDR {
				found = true
				break
			}
		}
		if !found {
			toAdd = append(toAdd, node)
		}
	}
	for _, r := range toDel {
		if dstRange != "" && !netutils.SubnetContainSubnet(dstRange, r) {
			continue
		}
		logrus.Info("delete route ", r)
		_, cidr, _ := net.ParseCIDR(r)
		if err = netlink.RouteDel(&netlink.Route{Dst: cidr}); err != nil {
			logrus.Error("failed to del route %v", err)
		}
	}

	for _, r := range toAdd {
		logrus.Info("add route ", r.ContainerCIDR, "via ", r.IP, "dev ", link)
		_, cidr, _ := net.ParseCIDR(r.ContainerCIDR)
		gw := net.ParseIP(r.IP)
		if err = netlink.RouteReplace(&netlink.Route{Dst: cidr, LinkIndex: nic.Attrs().Index, Scope: netlink.SCOPE_UNIVERSE, Gw: gw}); err != nil {
			logrus.Error("failed to add route %v", err)
		}
	}

	return
}
