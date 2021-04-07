package netutils

import (
	"fmt"
	"math/big"
	"net"
)

func InetNtoA(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func InetAtoN(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

func FirstSubnetIP(subnet string) (string, error) {
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", fmt.Errorf("%s is not a valid cidr", subnet)
	}
	ipInt := InetAtoN(cidr.IP.String())
	return InetNtoA(ipInt + 1), nil
}

func SubnetContainIP(subnet, ip string) bool {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err == nil {
		return ipnet.Contains(net.ParseIP(ip))
	}
	return false
}

func SubnetContainSubnet(src, dst string) bool {
	_, sipnet, err := net.ParseCIDR(src)
	if err != nil {
		return false
	}
	_, dipnet, err := net.ParseCIDR(dst)
	if err != nil {
		return false
	}
	return sipnet.Contains(dipnet.IP)
}
