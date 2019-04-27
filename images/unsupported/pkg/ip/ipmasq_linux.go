package ip

import (
	"fmt"
	"net"
	"github.com/coreos/go-iptables/iptables"
)

func SetupIPMasq(ipn *net.IPNet, chain string, comment string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	isV6 := ipn.IP.To4() == nil
	var ipt *iptables.IPTables
	var err error
	var multicastNet string
	if isV6 {
		ipt, err = iptables.NewWithProtocol(iptables.ProtocolIPv6)
		multicastNet = "ff00::/8"
	} else {
		ipt, err = iptables.NewWithProtocol(iptables.ProtocolIPv4)
		multicastNet = "224.0.0.0/4"
	}
	if err != nil {
		return fmt.Errorf("failed to locate iptables: %v", err)
	}
	exists := false
	chains, err := ipt.ListChains("nat")
	if err != nil {
		return fmt.Errorf("failed to list chains: %v", err)
	}
	for _, ch := range chains {
		if ch == chain {
			exists = true
			break
		}
	}
	if !exists {
		if err = ipt.NewChain("nat", chain); err != nil {
			return err
		}
	}
	if err := ipt.AppendUnique("nat", chain, "-d", ipn.String(), "-j", "ACCEPT", "-m", "comment", "--comment", comment); err != nil {
		return err
	}
	if err := ipt.AppendUnique("nat", chain, "!", "-d", multicastNet, "-j", "MASQUERADE", "-m", "comment", "--comment", comment); err != nil {
		return err
	}
	return ipt.AppendUnique("nat", "POSTROUTING", "-s", ipn.String(), "-j", chain, "-m", "comment", "--comment", comment)
}
func TeardownIPMasq(ipn *net.IPNet, chain string, comment string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	isV6 := ipn.IP.To4() == nil
	var ipt *iptables.IPTables
	var err error
	if isV6 {
		ipt, err = iptables.NewWithProtocol(iptables.ProtocolIPv6)
	} else {
		ipt, err = iptables.NewWithProtocol(iptables.ProtocolIPv4)
	}
	if err != nil {
		return fmt.Errorf("failed to locate iptables: %v", err)
	}
	if err = ipt.Delete("nat", "POSTROUTING", "-s", ipn.String(), "-j", chain, "-m", "comment", "--comment", comment); err != nil {
		return err
	}
	if err = ipt.ClearChain("nat", chain); err != nil {
		return err
	}
	return ipt.DeleteChain("nat", chain)
}
