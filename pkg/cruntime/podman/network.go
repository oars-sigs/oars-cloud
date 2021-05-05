package podman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/oars-sigs/oars-cloud/pkg/utils/netutils"
)

func (c *client) CreateNetwork(ctx context.Context, name, driver, subnet string) error {
	gateway, err := netutils.FirstSubnetIP(subnet)
	if err != nil {
		return err
	}
	ip := net.ParseIP(gateway)
	_, ipnet, err := net.ParseCIDR(subnet)
	req := networkCreateReq{
		Driver:  driver,
		Gateway: ip,
		Subnet:  ipnet,
	}
	jsonString, err := json.Marshal(req)
	if err != nil {
		return err
	}
	res, err := c.Post(ctx, "/libpod/network/create", bytes.NewBuffer(jsonString))
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("unknown error, status code: %d: %s", res.StatusCode, body)
	}
	return nil

}

type networkCreateReq struct {
	Driver  string
	Gateway net.IP
	Subnet  *net.IPNet
}

func (c *client) ListNetworks(ctx context.Context) ([]string, error) {
	return []string{}, nil
}
