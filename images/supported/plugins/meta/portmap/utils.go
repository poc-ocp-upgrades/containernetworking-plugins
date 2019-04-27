package main

import (
	"crypto/sha512"
	"fmt"
	"net"
	"strconv"
	"strings"
	"github.com/vishvananda/netlink"
)

const maxChainNameLength = 28

func fmtIpPort(ip net.IP, port int) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if ip.To4() == nil {
		return fmt.Sprintf("[%s]:%d", ip.String(), port)
	}
	return fmt.Sprintf("%s:%d", ip.String(), port)
}
func localhostIP(isV6 bool) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if isV6 {
		return "::1"
	}
	return "127.0.0.1"
}
func getRoutableHostIF(containerIP net.IP) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	routes, err := netlink.RouteGet(containerIP)
	if err != nil {
		return ""
	}
	for _, route := range routes {
		link, err := netlink.LinkByIndex(route.LinkIndex)
		if err != nil {
			continue
		}
		return link.Attrs().Name
	}
	return ""
}
func formatChainName(prefix, name, id string) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	chainBytes := sha512.Sum512([]byte(name + id))
	chain := fmt.Sprintf("CNI-%s%x", prefix, chainBytes)
	return chain[:maxChainNameLength]
}
func groupByProto(entries []PortMapEntry) map[string][]int {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(entries) == 0 {
		return map[string][]int{}
	}
	out := map[string][]int{}
	for _, e := range entries {
		_, ok := out[e.Protocol]
		if ok {
			out[e.Protocol] = append(out[e.Protocol], e.HostPort)
		} else {
			out[e.Protocol] = []int{e.HostPort}
		}
	}
	return out
}
func splitPortList(l []int) []string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	out := []string{}
	acc := []string{}
	for _, i := range l {
		acc = append(acc, strconv.Itoa(i))
		if len(acc) == 15 {
			out = append(out, strings.Join(acc, ","))
			acc = []string{}
		}
	}
	if len(acc) > 0 {
		out = append(out, strings.Join(acc, ","))
	}
	return out
}
func trimComment(val string) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(val) <= 255 {
		return val
	}
	return val[0:253] + "..."
}
