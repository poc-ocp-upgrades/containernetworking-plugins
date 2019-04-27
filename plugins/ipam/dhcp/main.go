package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"path/filepath"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
)

const socketPath = "/run/cni/dhcp.sock"

func main() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(os.Args) > 1 && os.Args[1] == "daemon" {
		var pidfilePath string
		var hostPrefix string
		daemonFlags := flag.NewFlagSet("daemon", flag.ExitOnError)
		daemonFlags.StringVar(&pidfilePath, "pidfile", "", "optional path to write daemon PID to")
		daemonFlags.StringVar(&hostPrefix, "hostprefix", "", "optional prefix to netns")
		daemonFlags.Parse(os.Args[2:])
		if err := runDaemon(pidfilePath, hostPrefix); err != nil {
			log.Printf(err.Error())
			os.Exit(1)
		}
	} else {
		skel.PluginMain(cmdAdd, cmdDel, version.All)
	}
}
func cmdAdd(args *skel.CmdArgs) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	versionDecoder := &version.ConfigDecoder{}
	confVersion, err := versionDecoder.Decode(args.StdinData)
	if err != nil {
		return err
	}
	result := &current.Result{}
	if err := rpcCall("DHCP.Allocate", args, result); err != nil {
		return err
	}
	return types.PrintResult(result, confVersion)
}
func cmdDel(args *skel.CmdArgs) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	result := struct{}{}
	if err := rpcCall("DHCP.Release", args, &result); err != nil {
		return fmt.Errorf("error dialing DHCP daemon: %v", err)
	}
	return nil
}
func rpcCall(method string, args *skel.CmdArgs, result interface{}) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	client, err := rpc.DialHTTP("unix", socketPath)
	if err != nil {
		return fmt.Errorf("error dialing DHCP daemon: %v", err)
	}
	netns, err := filepath.Abs(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to make %q an absolute path: %v", args.Netns, err)
	}
	args.Netns = netns
	err = client.Call(method, args, result)
	if err != nil {
		return fmt.Errorf("error calling %v: %v", method, err)
	}
	return nil
}
