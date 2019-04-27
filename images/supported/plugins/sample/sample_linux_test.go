package main

import (
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/testutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("sample test", func() {
	var targetNs ns.NetNS
	BeforeEach(func() {
		var err error
		targetNs, err = ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		targetNs.Close()
	})
	It("Works with a 0.3.0 config", func() {
		ifname := "eth0"
		conf := `{
	"cniVersion": "0.3.0",
	"name": "cni-plugin-sample-test",
	"type": "sample",
	"anotherAwesomeArg": "awesome",
	"prevResult": {
		"interfaces": [
			{
				"name": "%s",
				"sandbox": "%s"
			}
		],
		"ips": [
			{
				"version": "4",
				"address": "10.0.0.2/24",
				"gateway": "10.0.0.1",
				"interface": 0
			}
		],
		"routes": []
	}
}`
		conf = fmt.Sprintf(conf, ifname, targetNs.Path())
		args := &skel.CmdArgs{ContainerID: "dummy", Netns: targetNs.Path(), IfName: ifname, StdinData: []byte(conf)}
		_, _, err := testutils.CmdAddWithResult(targetNs.Path(), "eth0", []byte(conf), func() error {
			return cmdAdd(args)
		})
		Expect(err).NotTo(HaveOccurred())
	})
	It("fails an invalid config", func() {
		conf := `{
	"cniVersion": "0.3.0",
	"name": "cni-plugin-sample-test",
	"type": "sample",
	"prevResult": {
		"interfaces": [
			{
				"name": "eth0",
				"sandbox": "/var/run/netns/test"
			}
		],
		"ips": [
			{
				"version": "4",
				"address": "10.0.0.2/24",
				"gateway": "10.0.0.1",
				"interface": 0
			}
		],
		"routes": []
	}
}`
		args := &skel.CmdArgs{ContainerID: "dummy", Netns: targetNs.Path(), IfName: "eth0", StdinData: []byte(conf)}
		_, _, err := testutils.CmdAddWithResult(targetNs.Path(), "eth0", []byte(conf), func() error {
			return cmdAdd(args)
		})
		Expect(err).To(MatchError("anotherAwesomeArg must be specified"))
	})
	It("works with a 0.2.0 config", func() {
		conf := `{
	"cniVersion": "0.2.0",
	"name": "cni-plugin-sample-test",
	"type": "sample",
	"anotherAwesomeArg": "foo",
	"prevResult": {
		"ip4": {
			"ip": "10.0.0.2/24",
			"gateway": "10.0.0.1",
			"routes": []
		}
	}
}`
		args := &skel.CmdArgs{ContainerID: "dummy", Netns: targetNs.Path(), IfName: "eth0", StdinData: []byte(conf)}
		_, _, err := testutils.CmdAddWithResult(targetNs.Path(), "eth0", []byte(conf), func() error {
			return cmdAdd(args)
		})
		Expect(err).NotTo(HaveOccurred())
	})
})
