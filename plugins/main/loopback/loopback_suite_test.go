package main_test

import (
	"github.com/onsi/gomega/gexec"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

var pathToLoPlugin string

func TestLoopback(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loopback Suite")
}

var _ = BeforeSuite(func() {
	var err error
	pathToLoPlugin, err = gexec.Build("github.com/containernetworking/plugins/plugins/main/loopback")
	Expect(err).NotTo(HaveOccurred())
})
var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
