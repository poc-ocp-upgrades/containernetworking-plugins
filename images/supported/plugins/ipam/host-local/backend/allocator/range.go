package allocator

import (
	"fmt"
	"net"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/plugins/pkg/ip"
)

func (r *Range) Canonicalize() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if err := canonicalizeIP(&r.Subnet.IP); err != nil {
		return err
	}
	ones, masklen := r.Subnet.Mask.Size()
	if ones > masklen-2 {
		return fmt.Errorf("Network %s too small to allocate from", (*net.IPNet)(&r.Subnet).String())
	}
	if len(r.Subnet.IP) != len(r.Subnet.Mask) {
		return fmt.Errorf("IPNet IP and Mask version mismatch")
	}
	if r.Gateway == nil {
		r.Gateway = ip.NextIP(r.Subnet.IP)
	} else {
		if err := canonicalizeIP(&r.Gateway); err != nil {
			return err
		}
		subnet := (net.IPNet)(r.Subnet)
		if !subnet.Contains(r.Gateway) {
			return fmt.Errorf("gateway %s not in network %s", r.Gateway.String(), subnet.String())
		}
	}
	if r.RangeStart != nil {
		if err := canonicalizeIP(&r.RangeStart); err != nil {
			return err
		}
		if !r.Contains(r.RangeStart) {
			return fmt.Errorf("RangeStart %s not in network %s", r.RangeStart.String(), (*net.IPNet)(&r.Subnet).String())
		}
	} else {
		r.RangeStart = ip.NextIP(r.Subnet.IP)
	}
	if r.RangeEnd != nil {
		if err := canonicalizeIP(&r.RangeEnd); err != nil {
			return err
		}
		if !r.Contains(r.RangeEnd) {
			return fmt.Errorf("RangeEnd %s not in network %s", r.RangeEnd.String(), (*net.IPNet)(&r.Subnet).String())
		}
	} else {
		r.RangeEnd = lastIP(r.Subnet)
	}
	return nil
}
func (r *Range) Contains(addr net.IP) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if err := canonicalizeIP(&addr); err != nil {
		return false
	}
	subnet := (net.IPNet)(r.Subnet)
	if len(addr) != len(r.Subnet.IP) {
		return false
	}
	if !subnet.Contains(addr) {
		return false
	}
	if r.RangeStart != nil {
		if ip.Cmp(addr, r.RangeStart) < 0 {
			return false
		}
	}
	if r.RangeEnd != nil {
		if ip.Cmp(addr, r.RangeEnd) > 0 {
			return false
		}
	}
	return true
}
func (r *Range) Overlaps(r1 *Range) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(r.RangeStart) != len(r1.RangeStart) {
		return false
	}
	return r.Contains(r1.RangeStart) || r.Contains(r1.RangeEnd) || r1.Contains(r.RangeStart) || r1.Contains(r.RangeEnd)
}
func (r *Range) String() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return fmt.Sprintf("%s-%s", r.RangeStart.String(), r.RangeEnd.String())
}
func canonicalizeIP(ip *net.IP) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if ip.To4() != nil {
		*ip = ip.To4()
		return nil
	} else if ip.To16() != nil {
		*ip = ip.To16()
		return nil
	}
	return fmt.Errorf("IP %s not v4 nor v6", *ip)
}
func lastIP(subnet types.IPNet) net.IP {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var end net.IP
	for i := 0; i < len(subnet.IP); i++ {
		end = append(end, subnet.IP[i]|^subnet.Mask[i])
	}
	if subnet.IP.To4() != nil {
		end[3]--
	}
	return end
}
