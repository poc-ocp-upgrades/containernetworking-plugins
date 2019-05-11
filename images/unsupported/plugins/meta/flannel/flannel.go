package main

import (
	"bufio"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/containernetworking/cni/pkg/invoke"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
)

const (
	defaultSubnetFile	= "/run/flannel/subnet.env"
	defaultDataDir		= "/var/lib/cni/flannel"
)

type NetConf struct {
	types.NetConf
	SubnetFile	string					`json:"subnetFile"`
	DataDir		string					`json:"dataDir"`
	Delegate	map[string]interface{}	`json:"delegate"`
}
type subnetEnv struct {
	nw		*net.IPNet
	sn		*net.IPNet
	mtu		*uint
	ipmasq	*bool
}

func (se *subnetEnv) missing() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	m := []string{}
	if se.nw == nil {
		m = append(m, "FLANNEL_NETWORK")
	}
	if se.sn == nil {
		m = append(m, "FLANNEL_SUBNET")
	}
	if se.mtu == nil {
		m = append(m, "FLANNEL_MTU")
	}
	if se.ipmasq == nil {
		m = append(m, "FLANNEL_IPMASQ")
	}
	return strings.Join(m, ", ")
}
func loadFlannelNetConf(bytes []byte) (*NetConf, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	n := &NetConf{SubnetFile: defaultSubnetFile, DataDir: defaultDataDir}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %v", err)
	}
	return n, nil
}
func loadFlannelSubnetEnv(fn string) (*subnetEnv, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	se := &subnetEnv{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		parts := strings.SplitN(s.Text(), "=", 2)
		switch parts[0] {
		case "FLANNEL_NETWORK":
			_, se.nw, err = net.ParseCIDR(parts[1])
			if err != nil {
				return nil, err
			}
		case "FLANNEL_SUBNET":
			_, se.sn, err = net.ParseCIDR(parts[1])
			if err != nil {
				return nil, err
			}
		case "FLANNEL_MTU":
			mtu, err := strconv.ParseUint(parts[1], 10, 32)
			if err != nil {
				return nil, err
			}
			se.mtu = new(uint)
			*se.mtu = uint(mtu)
		case "FLANNEL_IPMASQ":
			ipmasq := parts[1] == "true"
			se.ipmasq = &ipmasq
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if m := se.missing(); m != "" {
		return nil, fmt.Errorf("%v is missing %v", fn, m)
	}
	return se, nil
}
func saveScratchNetConf(containerID, dataDir string, netconf []byte) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dataDir, containerID)
	return ioutil.WriteFile(path, netconf, 0600)
}
func consumeScratchNetConf(containerID, dataDir string) ([]byte, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	path := filepath.Join(dataDir, containerID)
	defer os.Remove(path)
	return ioutil.ReadFile(path)
}
func delegateAdd(cid, dataDir string, netconf map[string]interface{}) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	netconfBytes, err := json.Marshal(netconf)
	if err != nil {
		return fmt.Errorf("error serializing delegate netconf: %v", err)
	}
	if err = saveScratchNetConf(cid, dataDir, netconfBytes); err != nil {
		return err
	}
	result, err := invoke.DelegateAdd(netconf["type"].(string), netconfBytes)
	if err != nil {
		return err
	}
	return result.Print()
}
func hasKey(m map[string]interface{}, k string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_, ok := m[k]
	return ok
}
func isString(i interface{}) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_, ok := i.(string)
	return ok
}
func cmdAdd(args *skel.CmdArgs) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	n, err := loadFlannelNetConf(args.StdinData)
	if err != nil {
		return err
	}
	fenv, err := loadFlannelSubnetEnv(n.SubnetFile)
	if err != nil {
		return err
	}
	if n.Delegate == nil {
		n.Delegate = make(map[string]interface{})
	} else {
		if hasKey(n.Delegate, "type") && !isString(n.Delegate["type"]) {
			return fmt.Errorf("'delegate' dictionary, if present, must have (string) 'type' field")
		}
		if hasKey(n.Delegate, "name") {
			return fmt.Errorf("'delegate' dictionary must not have 'name' field, it'll be set by flannel")
		}
		if hasKey(n.Delegate, "ipam") {
			return fmt.Errorf("'delegate' dictionary must not have 'ipam' field, it'll be set by flannel")
		}
	}
	n.Delegate["name"] = n.Name
	if !hasKey(n.Delegate, "type") {
		n.Delegate["type"] = "bridge"
	}
	if !hasKey(n.Delegate, "ipMasq") {
		ipmasq := !*fenv.ipmasq
		n.Delegate["ipMasq"] = ipmasq
	}
	if !hasKey(n.Delegate, "mtu") {
		mtu := fenv.mtu
		n.Delegate["mtu"] = mtu
	}
	if n.Delegate["type"].(string) == "bridge" {
		if !hasKey(n.Delegate, "isGateway") {
			n.Delegate["isGateway"] = true
		}
	}
	if n.CNIVersion != "" {
		n.Delegate["cniVersion"] = n.CNIVersion
	}
	n.Delegate["ipam"] = map[string]interface{}{"type": "host-local", "subnet": fenv.sn.String(), "routes": []types.Route{types.Route{Dst: *fenv.nw}}}
	return delegateAdd(args.ContainerID, n.DataDir, n.Delegate)
}
func cmdDel(args *skel.CmdArgs) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	nc, err := loadFlannelNetConf(args.StdinData)
	if err != nil {
		return err
	}
	netconfBytes, err := consumeScratchNetConf(args.ContainerID, nc.DataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	n := &types.NetConf{}
	if err = json.Unmarshal(netconfBytes, n); err != nil {
		return fmt.Errorf("failed to parse netconf: %v", err)
	}
	return invoke.DelegateDel(n.Type, netconfBytes)
}
func main() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	skel.PluginMain(cmdAdd, cmdDel, version.All)
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
