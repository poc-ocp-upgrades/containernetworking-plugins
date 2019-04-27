package main

import (
	"net"
	"testing"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/d2g/dhcp4"
)

func validateRoutes(t *testing.T, routes []*types.Route) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	expected := []*types.Route{&types.Route{Dst: net.IPNet{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)}, GW: net.IPv4(10, 1, 2, 3)}, &types.Route{Dst: net.IPNet{IP: net.IPv4(192, 168, 1, 0), Mask: net.CIDRMask(24, 32)}, GW: net.IPv4(192, 168, 2, 3)}}
	if len(routes) != len(expected) {
		t.Fatalf("wrong length slice; expected %v, got %v", len(expected), len(routes))
	}
	for i := 0; i < len(routes); i++ {
		a := routes[i]
		e := expected[i]
		if a.Dst.String() != e.Dst.String() {
			t.Errorf("route.Dst mismatch: expected %v, got %v", e.Dst, a.Dst)
		}
		if !a.GW.Equal(e.GW) {
			t.Errorf("route.GW mismatch: expected %v, got %v", e.GW, a.GW)
		}
	}
}
func TestParseRoutes(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	opts := make(dhcp4.Options)
	opts[dhcp4.OptionStaticRoute] = []byte{10, 0, 0, 0, 10, 1, 2, 3, 192, 168, 1, 0, 192, 168, 2, 3}
	routes := parseRoutes(opts)
	validateRoutes(t, routes)
}
func TestParseCIDRRoutes(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	opts := make(dhcp4.Options)
	opts[dhcp4.OptionClasslessRouteFormat] = []byte{8, 10, 10, 1, 2, 3, 24, 192, 168, 1, 192, 168, 2, 3}
	routes := parseCIDRRoutes(opts)
	validateRoutes(t, routes)
}
