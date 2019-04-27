package ns_test

import (
	"math/rand"
	"runtime"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"testing"
)

func TestNs(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	rand.Seed(config.GinkgoConfig.RandomSeed)
	runtime.LockOSThread()
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/ns Suite")
}
