package allocator

import (
	"encoding/json"
	"fmt"
	"net"
	"github.com/containernetworking/cni/pkg/types"
	types020 "github.com/containernetworking/cni/pkg/types/020"
)

type Net struct {
	Name			string		`json:"name"`
	CNIVersion		string		`json:"cniVersion"`
	IPAM			*IPAMConfig	`json:"ipam"`
	RuntimeConfig	struct {
		IPRanges []RangeSet `json:"ipRanges,omitempty"`
	}	`json:"runtimeConfig,omitempty"`
	Args	*struct {
		A *IPAMArgs `json:"cni"`
	}	`json:"args"`
}
type IPAMConfig struct {
	*Range
	Name		string
	Type		string			`json:"type"`
	Routes		[]*types.Route	`json:"routes"`
	DataDir		string			`json:"dataDir"`
	ResolvConf	string			`json:"resolvConf"`
	Ranges		[]RangeSet		`json:"ranges"`
	IPArgs		[]net.IP		`json:"-"`
}
type IPAMEnvArgs struct {
	types.CommonArgs
	IP	net.IP	`json:"ip,omitempty"`
}
type IPAMArgs struct {
	IPs []net.IP `json:"ips"`
}
type RangeSet []Range
type Range struct {
	RangeStart	net.IP		`json:"rangeStart,omitempty"`
	RangeEnd	net.IP		`json:"rangeEnd,omitempty"`
	Subnet		types.IPNet	`json:"subnet"`
	Gateway		net.IP		`json:"gateway,omitempty"`
}

func LoadIPAMConfig(bytes []byte, envArgs string) (*IPAMConfig, string, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	n := Net{}
	if err := json.Unmarshal(bytes, &n); err != nil {
		return nil, "", err
	}
	if n.IPAM == nil {
		return nil, "", fmt.Errorf("IPAM config missing 'ipam' key")
	}
	if envArgs != "" {
		e := IPAMEnvArgs{}
		err := types.LoadArgs(envArgs, &e)
		if err != nil {
			return nil, "", err
		}
		if e.IP != nil {
			n.IPAM.IPArgs = []net.IP{e.IP}
		}
	}
	if n.Args != nil && n.Args.A != nil && len(n.Args.A.IPs) != 0 {
		n.IPAM.IPArgs = append(n.IPAM.IPArgs, n.Args.A.IPs...)
	}
	for idx, _ := range n.IPAM.IPArgs {
		if err := canonicalizeIP(&n.IPAM.IPArgs[idx]); err != nil {
			return nil, "", fmt.Errorf("cannot understand ip: %v", err)
		}
	}
	if n.IPAM.Range != nil && n.IPAM.Range.Subnet.IP != nil {
		n.IPAM.Ranges = append([]RangeSet{{*n.IPAM.Range}}, n.IPAM.Ranges...)
	}
	n.IPAM.Range = nil
	if len(n.RuntimeConfig.IPRanges) > 0 {
		n.IPAM.Ranges = append(n.RuntimeConfig.IPRanges, n.IPAM.Ranges...)
	}
	if len(n.IPAM.Ranges) == 0 {
		return nil, "", fmt.Errorf("no IP ranges specified")
	}
	numV4 := 0
	numV6 := 0
	for i, _ := range n.IPAM.Ranges {
		if err := n.IPAM.Ranges[i].Canonicalize(); err != nil {
			return nil, "", fmt.Errorf("invalid range set %d: %s", i, err)
		}
		if n.IPAM.Ranges[i][0].RangeStart.To4() != nil {
			numV4++
		} else {
			numV6++
		}
	}
	if numV4 > 1 || numV6 > 1 {
		for _, v := range types020.SupportedVersions {
			if n.CNIVersion == v {
				return nil, "", fmt.Errorf("CNI version %v does not support more than 1 address per family", n.CNIVersion)
			}
		}
	}
	l := len(n.IPAM.Ranges)
	for i, p1 := range n.IPAM.Ranges[:l-1] {
		for j, p2 := range n.IPAM.Ranges[i+1:] {
			if p1.Overlaps(&p2) {
				return nil, "", fmt.Errorf("range set %d overlaps with %d", i, (i + j + 1))
			}
		}
	}
	n.IPAM.Name = n.Name
	return n.IPAM, n.CNIVersion, nil
}
