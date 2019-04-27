package main

import (
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/testutils"
	"github.com/vishvananda/netlink"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("tuning plugin", func() {
	var originalNS ns.NetNS
	const IFNAME string = "dummy0"
	BeforeEach(func() {
		var err error
		originalNS, err = ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			err = netlink.LinkAdd(&netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: IFNAME}})
			Expect(err).NotTo(HaveOccurred())
			_, err = netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		Expect(originalNS.Close()).To(Succeed())
	})
	It("passes prevResult through unchanged", func() {
		conf := []byte(`{
	"name": "test",
	"type": "tuning",
	"cniVersion": "0.3.1",
	"sysctl": {
		"net.ipv4.conf.all.log_martians": "1"
	},
	"prevResult": {
		"interfaces": [
			{"name": "dummy0", "sandbox":"netns"}
		],
		"ips": [
			{
				"version": "4",
				"address": "10.0.0.2/24",
				"gateway": "10.0.0.1",
				"interface": 0
			}
		]
	}
}`)
		targetNs, err := ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		defer targetNs.Close()
		args := &skel.CmdArgs{ContainerID: "dummy", Netns: targetNs.Path(), IfName: IFNAME, StdinData: conf}
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			r, _, err := testutils.CmdAddWithResult(targetNs.Path(), IFNAME, []byte(conf), func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())
			result, err := current.GetResult(r)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(result.Interfaces)).To(Equal(1))
			Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
			Expect(len(result.IPs)).To(Equal(1))
			Expect(result.IPs[0].Address.String()).To(Equal("10.0.0.2/24"))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})
})
