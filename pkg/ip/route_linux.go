package ip

import (
	"net"
	"github.com/vishvananda/netlink"
)

func AddRoute(ipn *net.IPNet, gw net.IP, dev netlink.Link) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return netlink.RouteAdd(&netlink.Route{LinkIndex: dev.Attrs().Index, Scope: netlink.SCOPE_UNIVERSE, Dst: ipn, Gw: gw})
}
func AddHostRoute(ipn *net.IPNet, gw net.IP, dev netlink.Link) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return netlink.RouteAdd(&netlink.Route{LinkIndex: dev.Attrs().Index, Scope: netlink.SCOPE_HOST, Dst: ipn, Gw: gw})
}
func AddDefaultRoute(gw net.IP, dev netlink.Link) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_, defNet, _ := net.ParseCIDR("0.0.0.0/0")
	return AddRoute(defNet, gw, dev)
}
