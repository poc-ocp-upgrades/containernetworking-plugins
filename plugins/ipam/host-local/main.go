package main

import (
	"fmt"
	"net"
	"strings"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/allocator"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/disk"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
)

func main() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	skel.PluginMain(cmdAdd, cmdDel, version.All)
}
func cmdAdd(args *skel.CmdArgs) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ipamConf, confVersion, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}
	result := &current.Result{}
	if ipamConf.ResolvConf != "" {
		dns, err := parseResolvConf(ipamConf.ResolvConf)
		if err != nil {
			return err
		}
		result.DNS = *dns
	}
	store, err := disk.New(ipamConf.Name, ipamConf.DataDir)
	if err != nil {
		return err
	}
	defer store.Close()
	allocs := []*allocator.IPAllocator{}
	requestedIPs := map[string]net.IP{}
	for _, ip := range ipamConf.IPArgs {
		requestedIPs[ip.String()] = ip
	}
	for idx, rangeset := range ipamConf.Ranges {
		allocator := allocator.NewIPAllocator(&rangeset, store, idx)
		var requestedIP net.IP
		for k, ip := range requestedIPs {
			if rangeset.Contains(ip) {
				requestedIP = ip
				delete(requestedIPs, k)
				break
			}
		}
		ipConf, err := allocator.Get(args.ContainerID, requestedIP)
		if err != nil {
			for _, alloc := range allocs {
				_ = alloc.Release(args.ContainerID)
			}
			return fmt.Errorf("failed to allocate for range %d: %v", idx, err)
		}
		allocs = append(allocs, allocator)
		result.IPs = append(result.IPs, ipConf)
	}
	if len(requestedIPs) != 0 {
		for _, alloc := range allocs {
			_ = alloc.Release(args.ContainerID)
		}
		errstr := "failed to allocate all requested IPs:"
		for _, ip := range requestedIPs {
			errstr = errstr + " " + ip.String()
		}
		return fmt.Errorf(errstr)
	}
	result.Routes = ipamConf.Routes
	return types.PrintResult(result, confVersion)
}
func cmdDel(args *skel.CmdArgs) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ipamConf, _, err := allocator.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}
	store, err := disk.New(ipamConf.Name, ipamConf.DataDir)
	if err != nil {
		return err
	}
	defer store.Close()
	var errors []string
	for idx, rangeset := range ipamConf.Ranges {
		ipAllocator := allocator.NewIPAllocator(&rangeset, store, idx)
		err := ipAllocator.Release(args.ContainerID)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}
	if errors != nil {
		return fmt.Errorf(strings.Join(errors, ";"))
	}
	return nil
}
