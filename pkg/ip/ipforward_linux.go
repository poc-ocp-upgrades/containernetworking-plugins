package ip

import (
	"bytes"
	"io/ioutil"
	"github.com/containernetworking/cni/pkg/types/current"
)

func EnableIP4Forward() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return echo1("/proc/sys/net/ipv4/ip_forward")
}
func EnableIP6Forward() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return echo1("/proc/sys/net/ipv6/conf/all/forwarding")
}
func EnableForward(ips []*current.IPConfig) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	v4 := false
	v6 := false
	for _, ip := range ips {
		if ip.Version == "4" && !v4 {
			if err := EnableIP4Forward(); err != nil {
				return err
			}
			v4 = true
		} else if ip.Version == "6" && !v6 {
			if err := EnableIP6Forward(); err != nil {
				return err
			}
			v6 = true
		}
	}
	return nil
}
func echo1(f string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if content, err := ioutil.ReadFile(f); err == nil {
		if bytes.Equal(bytes.TrimSpace(content), []byte("1")) {
			return nil
		}
	}
	return ioutil.WriteFile(f, []byte("1"), 0644)
}
