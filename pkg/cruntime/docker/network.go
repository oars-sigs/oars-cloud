package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"

	"github.com/oars-sigs/oars-cloud/pkg/utils/netutils"
)

func (d *daemon) CreateNetwork(ctx context.Context, name, driver, subnet string) error {
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
	_, err = d.client.NetworkCreate(ctx, name, nc)
	return err
}

func (d *daemon) ListNetworks(ctx context.Context) ([]string, error) {
	nets, err := d.client.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for _, n := range nets {
		res = append(res, n.Name)
	}
	return res, err
}
