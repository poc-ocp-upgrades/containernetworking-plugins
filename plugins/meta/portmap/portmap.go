package main

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"github.com/containernetworking/plugins/pkg/utils/sysctl"
	"github.com/coreos/go-iptables/iptables"
)

const TopLevelDNATChainName = "CNI-HOSTPORT-DNAT"
const SetMarkChainName = "CNI-HOSTPORT-SETMARK"
const MarkMasqChainName = "CNI-HOSTPORT-MASQ"
const OldTopLevelSNATChainName = "CNI-HOSTPORT-SNAT"

func forwardPorts(config *PortMapConf, containerIP net.IP) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	isV6 := (containerIP.To4() == nil)
	var ipt *iptables.IPTables
	var err error
	if isV6 {
		ipt, err = iptables.NewWithProtocol(iptables.ProtocolIPv6)
	} else {
		ipt, err = iptables.NewWithProtocol(iptables.ProtocolIPv4)
	}
	if err != nil {
		return fmt.Errorf("failed to open iptables: %v", err)
	}
	if *config.SNAT {
		if config.ExternalSetMarkChain == nil {
			setMarkChain := genSetMarkChain(*config.MarkMasqBit)
			if err := setMarkChain.setup(ipt); err != nil {
				return fmt.Errorf("unable to create chain %s: %v", setMarkChain.name, err)
			}
			masqChain := genMarkMasqChain(*config.MarkMasqBit)
			if err := masqChain.setup(ipt); err != nil {
				return fmt.Errorf("unable to create chain %s: %v", setMarkChain.name, err)
			}
		}
		if !isV6 {
			hostIfName := getRoutableHostIF(containerIP)
			if hostIfName != "" {
				if err := enableLocalnetRouting(hostIfName); err != nil {
					return fmt.Errorf("unable to enable route_localnet: %v", err)
				}
			}
		}
	}
	toplevelDnatChain := genToplevelDnatChain()
	if err := toplevelDnatChain.setup(ipt); err != nil {
		return fmt.Errorf("failed to create top-level DNAT chain: %v", err)
	}
	dnatChain := genDnatChain(config.Name, config.ContainerID)
	fillDnatRules(&dnatChain, config, containerIP)
	if err := dnatChain.setup(ipt); err != nil {
		return fmt.Errorf("unable to setup DNAT: %v", err)
	}
	return nil
}
func genToplevelDnatChain() chain {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return chain{table: "nat", name: TopLevelDNATChainName, entryRules: [][]string{{"-m", "addrtype", "--dst-type", "LOCAL"}}, entryChains: []string{"PREROUTING", "OUTPUT"}}
}
func genDnatChain(netName, containerID string) chain {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return chain{table: "nat", name: formatChainName("DN-", netName, containerID), entryChains: []string{TopLevelDNATChainName}}
}
func fillDnatRules(c *chain, config *PortMapConf, containerIP net.IP) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	isV6 := (containerIP.To4() == nil)
	comment := trimComment(fmt.Sprintf(`dnat name: "%s" id: "%s"`, config.Name, config.ContainerID))
	entries := config.RuntimeConfig.PortMaps
	setMarkChainName := SetMarkChainName
	if config.ExternalSetMarkChain != nil {
		setMarkChainName = *config.ExternalSetMarkChain
	}
	protoPorts := groupByProto(entries)
	protos := []string{}
	for proto := range protoPorts {
		protos = append(protos, proto)
	}
	sort.Strings(protos)
	for _, proto := range protos {
		for _, portSpec := range splitPortList(protoPorts[proto]) {
			r := []string{"-m", "comment", "--comment", comment, "-m", "multiport", "-p", proto, "--destination-ports", portSpec}
			if isV6 && config.ConditionsV6 != nil && len(*config.ConditionsV6) > 0 {
				r = append(r, *config.ConditionsV6...)
			} else if !isV6 && config.ConditionsV4 != nil && len(*config.ConditionsV4) > 0 {
				r = append(r, *config.ConditionsV4...)
			}
			c.entryRules = append(c.entryRules, r)
		}
	}
	c.rules = make([][]string, 0, 3*len(entries))
	for _, entry := range entries {
		ruleBase := []string{"-p", entry.Protocol, "--dport", strconv.Itoa(entry.HostPort)}
		if entry.HostIP != "" {
			ruleBase = append(ruleBase, "-d", entry.HostIP)
		}
		if *config.SNAT {
			hpRule := make([]string, len(ruleBase), len(ruleBase)+4)
			copy(hpRule, ruleBase)
			hpRule = append(hpRule, "-s", containerIP.String(), "-j", setMarkChainName)
			c.rules = append(c.rules, hpRule)
			if !isV6 {
				localRule := make([]string, len(ruleBase), len(ruleBase)+4)
				copy(localRule, ruleBase)
				localRule = append(localRule, "-s", "127.0.0.1", "-j", setMarkChainName)
				c.rules = append(c.rules, localRule)
			}
		}
		dnatRule := make([]string, len(ruleBase), len(ruleBase)+4)
		copy(dnatRule, ruleBase)
		dnatRule = append(dnatRule, "-j", "DNAT", "--to-destination", fmtIpPort(containerIP, entry.ContainerPort))
		c.rules = append(c.rules, dnatRule)
	}
}
func genSetMarkChain(markBit int) chain {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	markValue := 1 << uint(markBit)
	markDef := fmt.Sprintf("%#x/%#x", markValue, markValue)
	ch := chain{table: "nat", name: SetMarkChainName, rules: [][]string{{"-m", "comment", "--comment", "CNI portfwd masquerade mark", "-j", "MARK", "--set-xmark", markDef}}}
	return ch
}
func genMarkMasqChain(markBit int) chain {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	markValue := 1 << uint(markBit)
	markDef := fmt.Sprintf("%#x/%#x", markValue, markValue)
	ch := chain{table: "nat", name: MarkMasqChainName, entryChains: []string{"POSTROUTING"}, entryRules: [][]string{{"-m", "comment", "--comment", "CNI portfwd requiring masquerade"}}, rules: [][]string{{"-m", "mark", "--mark", markDef, "-j", "MASQUERADE"}}}
	return ch
}
func enableLocalnetRouting(ifName string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	routeLocalnetPath := "net.ipv4.conf." + ifName + ".route_localnet"
	_, err := sysctl.Sysctl(routeLocalnetPath, "1")
	return err
}
func genOldSnatChain(netName, containerID string) chain {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	name := formatChainName("SN-", netName, containerID)
	return chain{table: "nat", name: name, entryChains: []string{OldTopLevelSNATChainName}}
}
func unforwardPorts(config *PortMapConf) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	dnatChain := genDnatChain(config.Name, config.ContainerID)
	oldSnatChain := genOldSnatChain(config.Name, config.ContainerID)
	ip4t := maybeGetIptables(false)
	ip6t := maybeGetIptables(true)
	if ip4t == nil && ip6t == nil {
		return fmt.Errorf("neither iptables nor ip6tables usable")
	}
	if ip4t != nil {
		if err := dnatChain.teardown(ip4t); err != nil {
			return fmt.Errorf("could not teardown ipv4 dnat: %v", err)
		}
		oldSnatChain.teardown(ip4t)
	}
	if ip6t != nil {
		if err := dnatChain.teardown(ip6t); err != nil {
			return fmt.Errorf("could not teardown ipv6 dnat: %v", err)
		}
		oldSnatChain.teardown(ip6t)
	}
	return nil
}
func maybeGetIptables(isV6 bool) *iptables.IPTables {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	proto := iptables.ProtocolIPv4
	if isV6 {
		proto = iptables.ProtocolIPv6
	}
	ipt, err := iptables.NewWithProtocol(proto)
	if err != nil {
		return nil
	}
	_, err = ipt.List("nat", "OUTPUT")
	if err != nil {
		return nil
	}
	return ipt
}
