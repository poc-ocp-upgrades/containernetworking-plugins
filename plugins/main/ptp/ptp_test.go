package main

import (
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/testutils"
	"github.com/vishvananda/netlink"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ptp Operations", func() {
	var originalNS ns.NetNS
	BeforeEach(func() {
		var err error
		originalNS, err = ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		Expect(originalNS.Close()).To(Succeed())
	})
	doTest := func(conf string, numIPs int) {
		const IFNAME = "ptp0"
		targetNs, err := ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		defer targetNs.Close()
		args := &skel.CmdArgs{ContainerID: "dummy", Netns: targetNs.Path(), IfName: IFNAME, StdinData: []byte(conf)}
		var resI types.Result
		var res *current.Result
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			resI, _, err = testutils.CmdAddWithResult(targetNs.Path(), IFNAME, []byte(conf), func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		res, err = current.NewResultFromResult(resI)
		Expect(err).NotTo(HaveOccurred())
		seenIPs := 0
		wantMac := ""
		err = targetNs.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			wantMac = link.Attrs().HardwareAddr.String()
			for _, ipc := range res.IPs {
				if *ipc.Interface != 1 {
					continue
				}
				seenIPs += 1
				saddr := ipc.Address.IP.String()
				daddr := ipc.Gateway.String()
				fmt.Fprintln(GinkgoWriter, "ping", saddr, "->", daddr)
				if err := testutils.Ping(saddr, daddr, (ipc.Version == "6"), 30); err != nil {
					return fmt.Errorf("ping %s -> %s failed: %s", saddr, daddr, err)
				}
			}
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(seenIPs).To(Equal(numIPs))
		Expect(res.Interfaces).To(HaveLen(2))
		Expect(res.Interfaces[0].Name).To(HavePrefix("veth"))
		Expect(res.Interfaces[0].Mac).To(HaveLen(17))
		Expect(res.Interfaces[0].Sandbox).To(BeEmpty())
		Expect(res.Interfaces[1].Name).To(Equal(IFNAME))
		Expect(res.Interfaces[1].Mac).To(Equal(wantMac))
		Expect(res.Interfaces[1].Sandbox).To(Equal(targetNs.Path()))
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			err := testutils.CmdDelWithResult(targetNs.Path(), IFNAME, func() error {
				return cmdDel(args)
			})
			Expect(err).NotTo(HaveOccurred())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		err = targetNs.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName(IFNAME)
			Expect(err).To(HaveOccurred())
			Expect(link).To(BeNil())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	}
	It("configures and deconfigures a ptp link with ADD/DEL", func() {
		conf := `{
    "cniVersion": "0.3.1",
    "name": "mynet",
    "type": "ptp",
    "ipMasq": true,
    "mtu": 5000,
    "ipam": {
        "type": "host-local",
        "subnet": "10.1.2.0/24"
    }
}`
		doTest(conf, 1)
	})
	It("configures and deconfigures a dual-stack ptp link with ADD/DEL", func() {
		conf := `{
    "cniVersion": "0.3.1",
    "name": "mynet",
    "type": "ptp",
    "ipMasq": true,
    "mtu": 5000,
    "ipam": {
        "type": "host-local",
		"ranges": [
			[{ "subnet": "10.1.2.0/24"}],
			[{ "subnet": "2001:db8:1::0/66"}]
		]
    }
}`
		doTest(conf, 2)
	})
	It("deconfigures an unconfigured ptp link with DEL", func() {
		const IFNAME = "ptp0"
		conf := `{
    "cniVersion": "0.3.0",
    "name": "mynet",
    "type": "ptp",
    "ipMasq": true,
    "mtu": 5000,
    "ipam": {
        "type": "host-local",
        "subnet": "10.1.2.0/24"
    }
}`
		targetNs, err := ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		defer targetNs.Close()
		args := &skel.CmdArgs{ContainerID: "dummy", Netns: targetNs.Path(), IfName: IFNAME, StdinData: []byte(conf)}
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			err := testutils.CmdDelWithResult(targetNs.Path(), IFNAME, func() error {
				return cmdDel(args)
			})
			Expect(err).NotTo(HaveOccurred())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})
})
