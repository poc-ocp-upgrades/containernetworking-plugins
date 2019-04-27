package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"github.com/containernetworking/cni/libcni"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/coreos/go-iptables/iptables"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/vishvananda/netlink"
)

const TIMEOUT = 90

var _ = Describe("portmap integration tests", func() {
	var (
		configList	*libcni.NetworkConfigList
		cniConf		*libcni.CNIConfig
		targetNS	ns.NetNS
		containerPort	int
		session		*gexec.Session
	)
	BeforeEach(func() {
		var err error
		rawConfig := `{
	"cniVersion": "0.3.0",
	"name": "cni-portmap-unit-test",
	"plugins": [
		{
			"type": "ptp",
			"ipMasq": true,
			"ipam": {
				"type": "host-local",
				"subnet": "172.16.31.0/24",
				"routes": [
					{"dst": "0.0.0.0/0"}
				]
			}
		},
		{
			"type": "portmap",
			"capabilities": {
				"portMappings": true
			}
		}
	]
}`
		configList, err = libcni.ConfListFromBytes([]byte(rawConfig))
		Expect(err).NotTo(HaveOccurred())
		dirs := filepath.SplitList(os.Getenv("PATH"))
		cniConf = &libcni.CNIConfig{Path: dirs}
		targetNS, err = ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		fmt.Fprintln(GinkgoWriter, "namespace:", targetNS.Path())
		containerPort, session, err = StartEchoServerInNamespace(targetNS)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		session.Terminate().Wait()
		if targetNS != nil {
			targetNS.Close()
		}
	})
	It("forwards a TCP port on ipv4", func(done Done) {
		var err error
		hostPort := rand.Intn(10000) + 1025
		runtimeConfig := libcni.RuntimeConf{ContainerID: fmt.Sprintf("unit-test-%d", hostPort), NetNS: targetNS.Path(), IfName: "eth0", CapabilityArgs: map[string]interface{}{"portMappings": []map[string]interface{}{{"hostPort": hostPort, "containerPort": containerPort, "protocol": "tcp"}}}}
		netDeleted := false
		deleteNetwork := func() error {
			if netDeleted {
				return nil
			}
			netDeleted = true
			return cniConf.DelNetworkList(configList, &runtimeConfig)
		}
		ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
		Expect(err).NotTo(HaveOccurred())
		dnatChainName := genDnatChain("cni-portmap-unit-test", runtimeConfig.ContainerID).name
		resI, err := cniConf.AddNetworkList(configList, &runtimeConfig)
		Expect(err).NotTo(HaveOccurred())
		defer deleteNetwork()
		cmd := exec.Command("iptables", "-t", "filter", "-P", "FORWARD", "ACCEPT")
		cmd.Stderr = GinkgoWriter
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred())
		_, err = ipt.List("nat", dnatChainName)
		Expect(err).NotTo(HaveOccurred())
		result, err := current.GetResult(resI)
		Expect(err).NotTo(HaveOccurred())
		var contIP net.IP
		for _, ip := range result.IPs {
			intfIndex := *ip.Interface
			if result.Interfaces[intfIndex].Sandbox == "" {
				continue
			}
			contIP = ip.Address.IP
		}
		if contIP == nil {
			Fail("could not determine container IP")
		}
		hostIP := getLocalIP()
		fmt.Fprintf(GinkgoWriter, "hostIP: %s:%d, contIP: %s:%d\n", hostIP, hostPort, contIP, containerPort)
		contOK := testEchoServer(contIP.String(), containerPort, "")
		dnatOK := testEchoServer(hostIP, hostPort, "")
		snatOK := testEchoServer("127.0.0.1", hostPort, "")
		hairpinOK := testEchoServer(hostIP, hostPort, targetNS.Path())
		session.Terminate()
		err = deleteNetwork()
		Expect(err).NotTo(HaveOccurred())
		_, err = ipt.List("nat", dnatChainName)
		Expect(err).To(MatchError(ContainSubstring("iptables: No chain/target/match by that name.")))
		if !contOK {
			Fail("connection direct to " + contIP.String() + " failed")
		}
		if !dnatOK {
			Fail("Connection to " + hostIP + " was not forwarded")
		}
		if !snatOK {
			Fail("connection to 127.0.0.1 was not forwarded")
		}
		if !hairpinOK {
			Fail("Hairpin connection failed")
		}
		close(done)
	}, TIMEOUT*9)
})

func testEchoServer(address string, port int, netns string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	message := "Aliquid melius quam pessimum optimum non est."
	bin, err := exec.LookPath("nc")
	Expect(err).NotTo(HaveOccurred())
	var cmd *exec.Cmd
	if netns != "" {
		netns = filepath.Base(netns)
		cmd = exec.Command("ip", "netns", "exec", netns, bin, "-v", address, strconv.Itoa(port))
	} else {
		cmd = exec.Command("nc", address, strconv.Itoa(port))
	}
	cmd.Stdin = bytes.NewBufferString(message)
	cmd.Stderr = GinkgoWriter
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(GinkgoWriter, "got non-zero exit from ", cmd.Args)
		return false
	}
	if string(out) != message {
		fmt.Fprintln(GinkgoWriter, "returned message didn't match?")
		fmt.Fprintln(GinkgoWriter, string(out))
		return false
	}
	return true
}
func getLocalIP() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	addrs, err := netlink.AddrList(nil, netlink.FAMILY_V4)
	Expect(err).NotTo(HaveOccurred())
	for _, addr := range addrs {
		if !addr.IP.IsGlobalUnicast() {
			continue
		}
		return addr.IP.String()
	}
	Fail("no live addresses")
	return ""
}
