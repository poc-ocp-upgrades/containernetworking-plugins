package disk

import (
	"github.com/alexflint/go-filemutex"
	"os"
	"path"
)

type FileLock struct{ f *filemutex.FileMutex }

func NewFileLock(lockPath string) (*FileLock, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	fi, err := os.Stat(lockPath)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		lockPath = path.Join(lockPath, "lock")
	}
	f, err := filemutex.New(lockPath)
	if err != nil {
		return nil, err
	}
	return &FileLock{f}, nil
}
func (l *FileLock) Close() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return l.f.Close()
}
func (l *FileLock) Lock() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return l.f.Lock()
}
func (l *FileLock) Unlock() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return l.f.Unlock()
}
