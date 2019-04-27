package main

import (
	"math/rand"
	"net"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/containernetworking/plugins/pkg/ns"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"testing"
)

func TestPortmap(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	rand.Seed(config.GinkgoConfig.RandomSeed)
	RegisterFailHandler(Fail)
	RunSpecs(t, "portmap Suite")
}

var echoServerBinaryPath string
var _ = SynchronizedBeforeSuite(func() []byte {
	binaryPath, err := gexec.Build("github.com/containernetworking/plugins/pkg/testutils/echosvr")
	Expect(err).NotTo(HaveOccurred())
	return []byte(binaryPath)
}, func(data []byte) {
	echoServerBinaryPath = string(data)
})
var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

func startInNetNS(binPath string, netNS ns.NetNS) (*gexec.Session, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	baseName := filepath.Base(netNS.Path())
	cmd := exec.Command("ip", "netns", "exec", baseName, binPath)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	return session, err
}
func StartEchoServerInNamespace(netNS ns.NetNS) (int, *gexec.Session, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	session, err := startInNetNS(echoServerBinaryPath, netNS)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session.Out).Should(gbytes.Say("\n"))
	_, portString, err := net.SplitHostPort(strings.TrimSpace(string(session.Out.Contents())))
	Expect(err).NotTo(HaveOccurred())
	port, err := strconv.Atoi(portString)
	Expect(err).NotTo(HaveOccurred())
	return port, session, nil
}
