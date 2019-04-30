package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/d2g/dhcp4"
)

func parseRouter(opts dhcp4.Options) net.IP {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if opts, ok := opts[dhcp4.OptionRouter]; ok {
		if len(opts) == 4 {
			return net.IP(opts)
		}
	}
	return nil
}
func classfulSubnet(sn net.IP) net.IPNet {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return net.IPNet{IP: sn, Mask: sn.DefaultMask()}
}
func parseRoutes(opts dhcp4.Options) []*types.Route {
	_logClusterCodePath()
	defer _logClusterCodePath()
	routes := []*types.Route{}
	if opt, ok := opts[dhcp4.OptionStaticRoute]; ok {
		for len(opt) >= 8 {
			sn := opt[0:4]
			r := opt[4:8]
			rt := &types.Route{Dst: classfulSubnet(sn), GW: r}
			routes = append(routes, rt)
			opt = opt[8:]
		}
	}
	return routes
}
func parseCIDRRoutes(opts dhcp4.Options) []*types.Route {
	_logClusterCodePath()
	defer _logClusterCodePath()
	routes := []*types.Route{}
	if opt, ok := opts[dhcp4.OptionClasslessRouteFormat]; ok {
		for len(opt) >= 5 {
			width := int(opt[0])
			if width > 32 {
				return nil
			}
			octets := 0
			if width > 0 {
				octets = (width-1)/8 + 1
			}
			if len(opt) < 1+octets+4 {
				return nil
			}
			sn := make([]byte, 4)
			copy(sn, opt[1:octets+1])
			gw := net.IP(opt[octets+1 : octets+5])
			rt := &types.Route{Dst: net.IPNet{IP: net.IP(sn), Mask: net.CIDRMask(width, 32)}, GW: gw}
			routes = append(routes, rt)
			opt = opt[octets+5 : len(opt)]
		}
	}
	return routes
}
func parseSubnetMask(opts dhcp4.Options) net.IPMask {
	_logClusterCodePath()
	defer _logClusterCodePath()
	mask, ok := opts[dhcp4.OptionSubnetMask]
	if !ok {
		return nil
	}
	return net.IPMask(mask)
}
func parseDuration(opts dhcp4.Options, code dhcp4.OptionCode, optName string) (time.Duration, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	val, ok := opts[code]
	if !ok {
		return 0, fmt.Errorf("option %v not found", optName)
	}
	if len(val) != 4 {
		return 0, fmt.Errorf("option %v is not 4 bytes", optName)
	}
	secs := binary.BigEndian.Uint32(val)
	return time.Duration(secs) * time.Second, nil
}
func parseLeaseTime(opts dhcp4.Options) (time.Duration, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return parseDuration(opts, dhcp4.OptionIPAddressLeaseTime, "LeaseTime")
}
func parseRenewalTime(opts dhcp4.Options) (time.Duration, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return parseDuration(opts, dhcp4.OptionRenewalTimeValue, "RenewalTime")
}
func parseRebindingTime(opts dhcp4.Options) (time.Duration, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return parseDuration(opts, dhcp4.OptionRebindingTimeValue, "RebindingTime")
}
