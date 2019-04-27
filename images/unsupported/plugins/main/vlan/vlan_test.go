package main

import (
	"fmt"
	"net"
	"syscall"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/testutils"
	"github.com/vishvananda/netlink"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const MASTER_NAME = "eth0"

var _ = Describe("vlan Operations", func() {
	var originalNS ns.NetNS
	BeforeEach(func() {
		var err error
		originalNS, err = ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			err = netlink.LinkAdd(&netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: MASTER_NAME}})
			Expect(err).NotTo(HaveOccurred())
			m, err := netlink.LinkByName(MASTER_NAME)
			Expect(err).NotTo(HaveOccurred())
			err = netlink.LinkSetUp(m)
			Expect(err).NotTo(HaveOccurred())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		Expect(originalNS.Close()).To(Succeed())
	})
	It("creates an vlan link in a non-default namespace with given MTU", func() {
		conf := &NetConf{NetConf: types.NetConf{CNIVersion: "0.3.0", Name: "testConfig", Type: "vlan"}, Master: MASTER_NAME, VlanId: 33, MTU: 1500}
		targetNs, err := ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		defer targetNs.Close()
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			_, err := createVlan(conf, "foobar0", targetNs)
			Expect(err).NotTo(HaveOccurred())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		err = targetNs.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName("foobar0")
			Expect(err).NotTo(HaveOccurred())
			Expect(link.Attrs().Name).To(Equal("foobar0"))
			Expect(link.Attrs().MTU).To(Equal(1500))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})
	It("creates an vlan link in a non-default namespace with master's MTU", func() {
		conf := &NetConf{NetConf: types.NetConf{CNIVersion: "0.3.0", Name: "testConfig", Type: "vlan"}, Master: MASTER_NAME, VlanId: 33}
		targetNs, err := ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		defer targetNs.Close()
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			m, err := netlink.LinkByName(MASTER_NAME)
			Expect(err).NotTo(HaveOccurred())
			err = netlink.LinkSetMTU(m, 1200)
			Expect(err).NotTo(HaveOccurred())
			_, err = createVlan(conf, "foobar0", targetNs)
			Expect(err).NotTo(HaveOccurred())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		err = targetNs.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName("foobar0")
			Expect(err).NotTo(HaveOccurred())
			Expect(link.Attrs().Name).To(Equal("foobar0"))
			Expect(link.Attrs().MTU).To(Equal(1200))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})
	It("configures and deconfigures an vlan link with ADD/DEL", func() {
		const IFNAME = "eth0"
		conf := fmt.Sprintf(`{
    "cniVersion": "0.3.0",
    "name": "mynet",
    "type": "vlan",
    "master": "%s",
    "ipam": {
        "type": "host-local",
        "subnet": "10.1.2.0/24"
    }
}`, MASTER_NAME)
		targetNs, err := ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		defer targetNs.Close()
		args := &skel.CmdArgs{ContainerID: "dummy", Netns: targetNs.Path(), IfName: IFNAME, StdinData: []byte(conf)}
		var result *current.Result
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			r, _, err := testutils.CmdAddWithResult(targetNs.Path(), IFNAME, []byte(conf), func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())
			result, err = current.GetResult(r)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(result.Interfaces)).To(Equal(1))
			Expect(result.Interfaces[0].Name).To(Equal(IFNAME))
			Expect(len(result.IPs)).To(Equal(1))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		err = targetNs.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			link, err := netlink.LinkByName(IFNAME)
			Expect(err).NotTo(HaveOccurred())
			Expect(link.Attrs().Name).To(Equal(IFNAME))
			hwaddr, err := net.ParseMAC(result.Interfaces[0].Mac)
			Expect(err).NotTo(HaveOccurred())
			Expect(link.Attrs().HardwareAddr).To(Equal(hwaddr))
			addrs, err := netlink.AddrList(link, syscall.AF_INET)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(addrs)).To(Equal(1))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		err = originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			err = testutils.CmdDelWithResult(targetNs.Path(), IFNAME, func() error {
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
	})
})
