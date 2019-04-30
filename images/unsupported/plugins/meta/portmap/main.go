package main

import (
	"encoding/json"
	"fmt"
	"net"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
)

type PortMapEntry struct {
	HostPort	int	`json:"hostPort"`
	ContainerPort	int	`json:"containerPort"`
	Protocol	string	`json:"protocol"`
	HostIP		string	`json:"hostIP,omitempty"`
}
type PortMapConf struct {
	types.NetConf
	SNAT			*bool		`json:"snat,omitempty"`
	ConditionsV4		*[]string	`json:"conditionsV4"`
	ConditionsV6		*[]string	`json:"conditionsV6"`
	MarkMasqBit		*int		`json:"markMasqBit"`
	ExternalSetMarkChain	*string		`json:"externalSetMarkChain"`
	RuntimeConfig		struct {
		PortMaps []PortMapEntry `json:"portMappings,omitempty"`
	}	`json:"runtimeConfig,omitempty"`
	RawPrevResult	map[string]interface{}	`json:"prevResult,omitempty"`
	PrevResult	*current.Result		`json:"-"`
	ContainerID	string			`json:"-"`
	ContIPv4	net.IP			`json:"-"`
	ContIPv6	net.IP			`json:"-"`
}

const DefaultMarkBit = 13

func cmdAdd(args *skel.CmdArgs) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	netConf, err := parseConfig(args.StdinData, args.IfName)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}
	if netConf.PrevResult == nil {
		return fmt.Errorf("must be called as chained plugin")
	}
	if len(netConf.RuntimeConfig.PortMaps) == 0 {
		return types.PrintResult(netConf.PrevResult, netConf.CNIVersion)
	}
	netConf.ContainerID = args.ContainerID
	if netConf.ContIPv4 != nil {
		if err := forwardPorts(netConf, netConf.ContIPv4); err != nil {
			return err
		}
	}
	if netConf.ContIPv6 != nil {
		if err := forwardPorts(netConf, netConf.ContIPv6); err != nil {
			return err
		}
	}
	return types.PrintResult(netConf.PrevResult, netConf.CNIVersion)
}
func cmdDel(args *skel.CmdArgs) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	netConf, err := parseConfig(args.StdinData, args.IfName)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}
	netConf.ContainerID = args.ContainerID
	if err := unforwardPorts(netConf); err != nil {
		return err
	}
	return nil
}
func main() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	skel.PluginMain(cmdAdd, cmdDel, version.PluginSupports("", "0.1.0", "0.2.0", "0.3.0", version.Current()))
}
func parseConfig(stdin []byte, ifName string) (*PortMapConf, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	conf := PortMapConf{}
	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse network configuration: %v", err)
	}
	if conf.RawPrevResult != nil {
		resultBytes, err := json.Marshal(conf.RawPrevResult)
		if err != nil {
			return nil, fmt.Errorf("could not serialize prevResult: %v", err)
		}
		res, err := version.NewResult(conf.CNIVersion, resultBytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse prevResult: %v", err)
		}
		conf.RawPrevResult = nil
		conf.PrevResult, err = current.NewResultFromResult(res)
		if err != nil {
			return nil, fmt.Errorf("could not convert result to current version: %v", err)
		}
	}
	if conf.SNAT == nil {
		tvar := true
		conf.SNAT = &tvar
	}
	if conf.MarkMasqBit != nil && conf.ExternalSetMarkChain != nil {
		return nil, fmt.Errorf("Cannot specify externalSetMarkChain and markMasqBit")
	}
	if conf.MarkMasqBit == nil {
		bvar := DefaultMarkBit
		conf.MarkMasqBit = &bvar
	}
	if *conf.MarkMasqBit < 0 || *conf.MarkMasqBit > 31 {
		return nil, fmt.Errorf("MasqMarkBit must be between 0 and 31")
	}
	for _, pm := range conf.RuntimeConfig.PortMaps {
		if pm.ContainerPort <= 0 {
			return nil, fmt.Errorf("Invalid container port number: %d", pm.ContainerPort)
		}
		if pm.HostPort <= 0 {
			return nil, fmt.Errorf("Invalid host port number: %d", pm.HostPort)
		}
	}
	if conf.PrevResult != nil {
		for _, ip := range conf.PrevResult.IPs {
			if ip.Version == "6" && conf.ContIPv6 != nil {
				continue
			} else if ip.Version == "4" && conf.ContIPv4 != nil {
				continue
			}
			if ip.Interface != nil {
				intIdx := *ip.Interface
				if intIdx >= 0 && intIdx < len(conf.PrevResult.Interfaces) && (conf.PrevResult.Interfaces[intIdx].Name != ifName || conf.PrevResult.Interfaces[intIdx].Sandbox == "") {
					continue
				}
			}
			switch ip.Version {
			case "6":
				conf.ContIPv6 = ip.Address.IP
			case "4":
				conf.ContIPv4 = ip.Address.IP
			}
		}
	}
	return &conf, nil
}
