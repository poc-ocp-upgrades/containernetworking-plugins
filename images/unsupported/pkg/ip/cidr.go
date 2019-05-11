package ip

import (
	"math/big"
	"net"
)

func NextIP(ip net.IP) net.IP {
	_logClusterCodePath()
	defer _logClusterCodePath()
	i := ipToInt(ip)
	return intToIP(i.Add(i, big.NewInt(1)))
}
func PrevIP(ip net.IP) net.IP {
	_logClusterCodePath()
	defer _logClusterCodePath()
	i := ipToInt(ip)
	return intToIP(i.Sub(i, big.NewInt(1)))
}
func Cmp(a, b net.IP) int {
	_logClusterCodePath()
	defer _logClusterCodePath()
	aa := ipToInt(a)
	bb := ipToInt(b)
	return aa.Cmp(bb)
}
func ipToInt(ip net.IP) *big.Int {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if v := ip.To4(); v != nil {
		return big.NewInt(0).SetBytes(v)
	}
	return big.NewInt(0).SetBytes(ip.To16())
}
func intToIP(i *big.Int) net.IP {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return net.IP(i.Bytes())
}
func Network(ipn *net.IPNet) *net.IPNet {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &net.IPNet{IP: ipn.IP.Mask(ipn.Mask), Mask: ipn.Mask}
}
