package testutils

import (
	"io/ioutil"
	"os"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
)

func envCleanup() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	os.Unsetenv("CNI_COMMAND")
	os.Unsetenv("CNI_PATH")
	os.Unsetenv("CNI_NETNS")
	os.Unsetenv("CNI_IFNAME")
}
func CmdAddWithResult(cniNetns, cniIfname string, conf []byte, f func() error) (types.Result, []byte, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	os.Setenv("CNI_COMMAND", "ADD")
	os.Setenv("CNI_PATH", os.Getenv("PATH"))
	os.Setenv("CNI_NETNS", cniNetns)
	os.Setenv("CNI_IFNAME", cniIfname)
	defer envCleanup()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	os.Stdout = w
	err = f()
	w.Close()
	var out []byte
	if err == nil {
		out, err = ioutil.ReadAll(r)
	}
	os.Stdout = oldStdout
	if err != nil {
		return nil, nil, err
	}
	versionDecoder := &version.ConfigDecoder{}
	confVersion, err := versionDecoder.Decode(conf)
	if err != nil {
		return nil, nil, err
	}
	result, err := version.NewResult(confVersion, out)
	if err != nil {
		return nil, nil, err
	}
	return result, out, nil
}
func CmdDelWithResult(cniNetns, cniIfname string, f func() error) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	os.Setenv("CNI_COMMAND", "DEL")
	os.Setenv("CNI_PATH", os.Getenv("PATH"))
	os.Setenv("CNI_NETNS", cniNetns)
	os.Setenv("CNI_IFNAME", cniIfname)
	defer envCleanup()
	return f()
}
