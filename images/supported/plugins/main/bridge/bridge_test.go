package main

import (
	"fmt"
	"net"
	"strings"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/020"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/testutils"
	"github.com/vishvananda/netlink"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	BRNAME	= "bridge0"
	IFNAME	= "eth0"
)

type testCase struct {
	cniVersion	string
	subnet		string
	gateway		string
	ranges		[]rangeInfo
	isGW		bool
	expGWCIDRs	[]string
}
type rangeInfo struct {
	subnet	string
	gateway	string
}

func (tc testCase) netConf() *NetConf {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &NetConf{NetConf: types.NetConf{CNIVersion: tc.cniVersion, Name: "testConfig", Type: "bridge"}, BrName: BRNAME, IsGW: tc.isGW, IPMasq: false, MTU: 5000}
}

const (
	netConfStr	= `
	"cniVersion": "%s",
	"name": "testConfig",
	"type": "bridge",
	"bridge": "%s",
	"isDefaultGateway": true,
	"ipMasq": false`
	ipamStartStr	= `,
    "ipam": {
        "type":    "host-local"`
	subnetConfStr	= `,
        "subnet":  "%s"`
	gatewayConfStr	= `,
        "gateway": "%s"`
	rangesStartStr	= `,
        "ranges": [`
	rangeSubnetConfStr	= `
            [{
                "subnet":  "%s"
            }]`
	rangeSubnetGWConfStr	= `
            [{
                "subnet":  "%s",
                "gateway": "%s"
            }]`
	rangesEndStr	= `
        ]`
	ipamEndStr	= `
    }`
)

func (tc testCase) netConfJSON() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	conf := fmt.Sprintf(netConfStr, tc.cniVersion, BRNAME)
	if tc.subnet != "" || tc.ranges != nil {
		conf += ipamStartStr
		if tc.subnet != "" {
			conf += tc.subnetConfig()
		}
		if tc.ranges != nil {
			conf += tc.rangesConfig()
		}
		conf += ipamEndStr
	}
	return "{" + conf + "\n}"
}
func (tc testCase) subnetConfig() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	conf := fmt.Sprintf(subnetConfStr, tc.subnet)
	if tc.gateway != "" {
		conf += fmt.Sprintf(gatewayConfStr, tc.gateway)
	}
	return conf
}
func (tc testCase) rangesConfig() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	conf := rangesStartStr
	for i, tcRange := range tc.ranges {
		if i > 0 {
			conf += ","
		}
		if tcRange.gateway != "" {
			conf += fmt.Sprintf(rangeSubnetGWConfStr, tcRange.subnet, tcRange.gateway)
		} else {
			conf += fmt.Sprintf(rangeSubnetConfStr, tcRange.subnet)
		}
	}
	return conf + rangesEndStr
}
func (tc testCase) createCmdArgs(targetNS ns.NetNS) *skel.CmdArgs {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	conf := tc.netConfJSON()
	return &skel.CmdArgs{ContainerID: "dummy", Netns: targetNS.Path(), IfName: IFNAME, StdinData: []byte(conf)}
}
func (tc testCase) expectedCIDRs() ([]*net.IPNet, []*net.IPNet) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var cidrsV4, cidrsV6 []*net.IPNet
	appendSubnet := func(subnet string) {
		ip, cidr, err := net.ParseCIDR(subnet)
		Expect(err).NotTo(HaveOccurred())
		if ipVersion(ip) == "4" {
			cidrsV4 = append(cidrsV4, cidr)
		} else {
			cidrsV6 = append(cidrsV6, cidr)
		}
	}
	if tc.subnet != "" {
		appendSubnet(tc.subnet)
	}
	for _, r := range tc.ranges {
		appendSubnet(r.subnet)
	}
	return cidrsV4, cidrsV6
}
func delBridgeAddrs(testNS ns.NetNS) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	err := testNS.Do(func(ns.NetNS) error {
		defer GinkgoRecover()
		br, err := netlink.LinkByName(BRNAME)
		Expect(err).NotTo(HaveOccurred())
		addrs, err := netlink.AddrList(br, netlink.FAMILY_ALL)
		Expect(err).NotTo(HaveOccurred())
		for _, addr := range addrs {
			if !addr.IP.IsLinkLocalUnicast() {
				err = netlink.AddrDel(br, &addr)
				Expect(err).NotTo(HaveOccurred())
			}
		}
		return nil
	})
	Expect(err).NotTo(HaveOccurred())
}
func ipVersion(ip net.IP) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if ip.To4() != nil {
		return "4"
	}
	return "6"
}

type cmdAddDelTester interface {
	setNS(testNS ns.NetNS, targetNS ns.NetNS)
	cmdAddTest(tc testCase)
	cmdDelTest(tc testCase)
}

func testerByVersion(version string) cmdAddDelTester {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch {
	case strings.HasPrefix(version, "0.3."):
		return &testerV03x{}
	default:
		return &testerV01xOr02x{}
	}
}

type testerV03x struct {
	testNS		ns.NetNS
	targetNS	ns.NetNS
	args		*skel.CmdArgs
	vethName	string
}

func (tester *testerV03x) setNS(testNS ns.NetNS, targetNS ns.NetNS) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	tester.testNS = testNS
	tester.targetNS = targetNS
}
func (tester *testerV03x) cmdAddTest(tc testCase) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	tester.args = tc.createCmdArgs(tester.targetNS)
	var result *current.Result
	err := tester.testNS.Do(func(ns.NetNS) error {
		defer GinkgoRecover()
		r, raw, err := testutils.CmdAddWithResult(tester.targetNS.Path(), IFNAME, tester.args.StdinData, func() error {
			return cmdAdd(tester.args)
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Index(string(raw), "\"interfaces\":")).Should(BeNumerically(">", 0))
		result, err = current.GetResult(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result.Interfaces)).To(Equal(3))
		Expect(result.Interfaces[0].Name).To(Equal(BRNAME))
		Expect(result.Interfaces[0].Mac).To(HaveLen(17))
		Expect(result.Interfaces[1].Name).To(HavePrefix("veth"))
		Expect(result.Interfaces[1].Mac).To(HaveLen(17))
		Expect(result.Interfaces[2].Name).To(Equal(IFNAME))
		Expect(result.Interfaces[2].Mac).To(HaveLen(17))
		Expect(result.Interfaces[2].Sandbox).To(Equal(tester.targetNS.Path()))
		link, err := netlink.LinkByName(result.Interfaces[0].Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(link.Attrs().Name).To(Equal(BRNAME))
		Expect(link).To(BeAssignableToTypeOf(&netlink.Bridge{}))
		Expect(link.Attrs().HardwareAddr.String()).To(Equal(result.Interfaces[0].Mac))
		bridgeMAC := link.Attrs().HardwareAddr.String()
		addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(addrs)).To(BeNumerically(">", 0))
		for _, cidr := range tc.expGWCIDRs {
			ip, subnet, err := net.ParseCIDR(cidr)
			Expect(err).NotTo(HaveOccurred())
			found := false
			subnetPrefix, subnetBits := subnet.Mask.Size()
			for _, a := range addrs {
				aPrefix, aBits := a.IPNet.Mask.Size()
				if a.IPNet.IP.Equal(ip) && aPrefix == subnetPrefix && aBits == subnetBits {
					found = true
					break
				}
			}
			Expect(found).To(Equal(true))
		}
		links, err := netlink.LinkList()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(links)).To(Equal(3))
		link, err = netlink.LinkByName(result.Interfaces[1].Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(link).To(BeAssignableToTypeOf(&netlink.Veth{}))
		tester.vethName = result.Interfaces[1].Name
		Expect(link.Attrs().HardwareAddr.String()).NotTo(Equal(bridgeMAC))
		return nil
	})
	Expect(err).NotTo(HaveOccurred())
	err = tester.targetNS.Do(func(ns.NetNS) error {
		defer GinkgoRecover()
		link, err := netlink.LinkByName(IFNAME)
		Expect(err).NotTo(HaveOccurred())
		Expect(link.Attrs().Name).To(Equal(IFNAME))
		Expect(link).To(BeAssignableToTypeOf(&netlink.Veth{}))
		expCIDRsV4, expCIDRsV6 := tc.expectedCIDRs()
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(addrs)).To(Equal(len(expCIDRsV4)))
		addrs, err = netlink.AddrList(link, netlink.FAMILY_V6)
		Expect(len(addrs)).To(Equal(len(expCIDRsV6) + 1))
		Expect(err).NotTo(HaveOccurred())
		var foundAddrs int
		for _, addr := range addrs {
			if !addr.IP.IsLinkLocalUnicast() {
				foundAddrs++
			}
		}
		Expect(foundAddrs).To(Equal(len(expCIDRsV6)))
		routes, err := netlink.RouteList(link, 0)
		Expect(err).NotTo(HaveOccurred())
		var defaultRouteFound4, defaultRouteFound6 bool
		for _, cidr := range tc.expGWCIDRs {
			gwIP, _, err := net.ParseCIDR(cidr)
			Expect(err).NotTo(HaveOccurred())
			var found *bool
			if ipVersion(gwIP) == "4" {
				found = &defaultRouteFound4
			} else {
				found = &defaultRouteFound6
			}
			if *found == true {
				continue
			}
			for _, route := range routes {
				*found = (route.Dst == nil && route.Src == nil && route.Gw.Equal(gwIP))
				if *found {
					break
				}
			}
			Expect(*found).To(Equal(true))
		}
		return nil
	})
	Expect(err).NotTo(HaveOccurred())
}
func (tester *testerV03x) cmdDelTest(tc testCase) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	err := tester.testNS.Do(func(ns.NetNS) error {
		defer GinkgoRecover()
		err := testutils.CmdDelWithResult(tester.targetNS.Path(), IFNAME, func() error {
			return cmdDel(tester.args)
		})
		Expect(err).NotTo(HaveOccurred())
		return nil
	})
	Expect(err).NotTo(HaveOccurred())
	err = tester.targetNS.Do(func(ns.NetNS) error {
		defer GinkgoRecover()
		link, err := netlink.LinkByName(IFNAME)
		Expect(err).To(HaveOccurred())
		Expect(link).To(BeNil())
		return nil
	})
	Expect(err).NotTo(HaveOccurred())
	err = tester.testNS.Do(func(ns.NetNS) error {
		defer GinkgoRecover()
		link, err := netlink.LinkByName(tester.vethName)
		Expect(err).To(HaveOccurred())
		Expect(link).To(BeNil())
		return nil
	})
}

type testerV01xOr02x struct {
	testNS		ns.NetNS
	targetNS	ns.NetNS
	args		*skel.CmdArgs
	vethName	string
}

func (tester *testerV01xOr02x) setNS(testNS ns.NetNS, targetNS ns.NetNS) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	tester.testNS = testNS
	tester.targetNS = targetNS
}
func (tester *testerV01xOr02x) cmdAddTest(tc testCase) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	tester.args = tc.createCmdArgs(tester.targetNS)
	var result *types020.Result
	err := tester.testNS.Do(func(ns.NetNS) error {
		defer GinkgoRecover()
		r, raw, err := testutils.CmdAddWithResult(tester.targetNS.Path(), IFNAME, tester.args.StdinData, func() error {
			return cmdAdd(tester.args)
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Index(string(raw), "\"ip\":")).Should(BeNumerically(">", 0))
		result, err = types020.GetResult(r)
		Expect(err).NotTo(HaveOccurred())
		link, err := netlink.LinkByName(BRNAME)
		Expect(err).NotTo(HaveOccurred())
		Expect(link.Attrs().Name).To(Equal(BRNAME))
		Expect(link).To(BeAssignableToTypeOf(&netlink.Bridge{}))
		addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(addrs)).To(BeNumerically(">", 0))
		for _, cidr := range tc.expGWCIDRs {
			ip, subnet, err := net.ParseCIDR(cidr)
			Expect(err).NotTo(HaveOccurred())
			found := false
			subnetPrefix, subnetBits := subnet.Mask.Size()
			for _, a := range addrs {
				aPrefix, aBits := a.IPNet.Mask.Size()
				if a.IPNet.IP.Equal(ip) && aPrefix == subnetPrefix && aBits == subnetBits {
					found = true
					break
				}
			}
			Expect(found).To(Equal(true))
		}
		links, err := netlink.LinkList()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(links)).To(Equal(3))
		return nil
	})
	Expect(err).NotTo(HaveOccurred())
	err = tester.targetNS.Do(func(ns.NetNS) error {
		defer GinkgoRecover()
		link, err := netlink.LinkByName(IFNAME)
		Expect(err).NotTo(HaveOccurred())
		Expect(link.Attrs().Name).To(Equal(IFNAME))
		Expect(link).To(BeAssignableToTypeOf(&netlink.Veth{}))
		expCIDRsV4, expCIDRsV6 := tc.expectedCIDRs()
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(addrs)).To(Equal(len(expCIDRsV4)))
		addrs, err = netlink.AddrList(link, netlink.FAMILY_V6)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(addrs)).To(Equal(len(expCIDRsV6) + 1))
		routes, err := netlink.RouteList(link, 0)
		Expect(err).NotTo(HaveOccurred())
		var defaultRouteFound bool
		for _, cidr := range tc.expGWCIDRs {
			gwIP, _, err := net.ParseCIDR(cidr)
			Expect(err).NotTo(HaveOccurred())
			for _, route := range routes {
				defaultRouteFound = (route.Dst == nil && route.Src == nil && route.Gw.Equal(gwIP))
				if defaultRouteFound {
					break
				}
			}
			Expect(defaultRouteFound).To(Equal(true))
		}
		return nil
	})
	Expect(err).NotTo(HaveOccurred())
}
func (tester *testerV01xOr02x) cmdDelTest(tc testCase) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	err := tester.testNS.Do(func(ns.NetNS) error {
		defer GinkgoRecover()
		err := testutils.CmdDelWithResult(tester.targetNS.Path(), IFNAME, func() error {
			return cmdDel(tester.args)
		})
		Expect(err).NotTo(HaveOccurred())
		return nil
	})
	Expect(err).NotTo(HaveOccurred())
	err = tester.testNS.Do(func(ns.NetNS) error {
		defer GinkgoRecover()
		link, err := netlink.LinkByName(IFNAME)
		Expect(err).To(HaveOccurred())
		Expect(link).To(BeNil())
		return nil
	})
	Expect(err).NotTo(HaveOccurred())
}
func cmdAddDelTest(testNS ns.NetNS, tc testCase) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	tester := testerByVersion(tc.cniVersion)
	targetNS, err := ns.NewNS()
	Expect(err).NotTo(HaveOccurred())
	defer targetNS.Close()
	tester.setNS(testNS, targetNS)
	tester.cmdAddTest(tc)
	tester.cmdDelTest(tc)
	delBridgeAddrs(testNS)
}

var _ = Describe("bridge Operations", func() {
	var originalNS ns.NetNS
	BeforeEach(func() {
		var err error
		originalNS, err = ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		Expect(originalNS.Close()).To(Succeed())
	})
	It("creates a bridge", func() {
		conf := testCase{cniVersion: "0.3.1"}.netConf()
		err := originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			bridge, _, err := setupBridge(conf)
			Expect(err).NotTo(HaveOccurred())
			Expect(bridge.Attrs().Name).To(Equal(BRNAME))
			link, err := netlink.LinkByName(BRNAME)
			Expect(err).NotTo(HaveOccurred())
			Expect(link.Attrs().Name).To(Equal(BRNAME))
			Expect(link.Attrs().Promisc).To(Equal(0))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})
	It("handles an existing bridge", func() {
		err := originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			err := netlink.LinkAdd(&netlink.Bridge{LinkAttrs: netlink.LinkAttrs{Name: BRNAME}})
			Expect(err).NotTo(HaveOccurred())
			link, err := netlink.LinkByName(BRNAME)
			Expect(err).NotTo(HaveOccurred())
			Expect(link.Attrs().Name).To(Equal(BRNAME))
			ifindex := link.Attrs().Index
			tc := testCase{cniVersion: "0.3.1", isGW: false}
			conf := tc.netConf()
			bridge, _, err := setupBridge(conf)
			Expect(err).NotTo(HaveOccurred())
			Expect(bridge.Attrs().Name).To(Equal(BRNAME))
			Expect(bridge.Attrs().Index).To(Equal(ifindex))
			link, err = netlink.LinkByName(BRNAME)
			Expect(err).NotTo(HaveOccurred())
			Expect(link.Attrs().Name).To(Equal(BRNAME))
			Expect(link.Attrs().Index).To(Equal(ifindex))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})
	It("configures and deconfigures a bridge and veth with default route with ADD/DEL for 0.3.0 config", func() {
		testCases := []testCase{{subnet: "10.1.2.0/24", expGWCIDRs: []string{"10.1.2.1/24"}}, {subnet: "2001:db8::0/64", expGWCIDRs: []string{"2001:db8::1/64"}}, {ranges: []rangeInfo{{subnet: "192.168.0.0/24"}, {subnet: "fd00::0/64"}}, expGWCIDRs: []string{"192.168.0.1/24", "fd00::1/64"}}, {ranges: []rangeInfo{{subnet: "192.168.0.0/24"}, {subnet: "fd00::0/64"}, {subnet: "2001:db8::0/64"}}, expGWCIDRs: []string{"192.168.0.1/24", "fd00::1/64", "2001:db8::1/64"}}}
		for _, tc := range testCases {
			tc.cniVersion = "0.3.0"
			cmdAddDelTest(originalNS, tc)
		}
	})
	It("configures and deconfigures a bridge and veth with default route with ADD/DEL for 0.3.1 config", func() {
		testCases := []testCase{{subnet: "10.1.2.0/24", expGWCIDRs: []string{"10.1.2.1/24"}}, {subnet: "2001:db8::0/64", expGWCIDRs: []string{"2001:db8::1/64"}}, {ranges: []rangeInfo{{subnet: "192.168.0.0/24"}, {subnet: "fd00::0/64"}}, expGWCIDRs: []string{"192.168.0.1/24", "fd00::1/64"}}}
		for _, tc := range testCases {
			tc.cniVersion = "0.3.1"
			cmdAddDelTest(originalNS, tc)
		}
	})
	It("deconfigures an unconfigured bridge with DEL", func() {
		tc := testCase{cniVersion: "0.3.0", subnet: "10.1.2.0/24", expGWCIDRs: []string{"10.1.2.1/24"}}
		tester := testerV03x{}
		targetNS, err := ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		defer targetNS.Close()
		tester.setNS(originalNS, targetNS)
		tester.args = tc.createCmdArgs(targetNS)
		tester.cmdDelTest(tc)
	})
	It("configures and deconfigures a bridge and veth with default route with ADD/DEL for 0.1.0 config", func() {
		testCases := []testCase{{subnet: "10.1.2.0/24", expGWCIDRs: []string{"10.1.2.1/24"}}, {subnet: "2001:db8::0/64", expGWCIDRs: []string{"2001:db8::1/64"}}, {ranges: []rangeInfo{{subnet: "192.168.0.0/24"}, {subnet: "fd00::0/64"}}, expGWCIDRs: []string{"192.168.0.1/24", "fd00::1/64"}}}
		for _, tc := range testCases {
			tc.cniVersion = "0.1.0"
			cmdAddDelTest(originalNS, tc)
		}
	})
	It("ensure bridge address", func() {
		conf := testCase{cniVersion: "0.3.1", isGW: true}.netConf()
		testCases := []struct {
			gwCIDRFirst	string
			gwCIDRSecond	string
		}{{gwCIDRFirst: "10.0.0.1/8", gwCIDRSecond: "10.1.2.3/16"}, {gwCIDRFirst: "2001:db8:1::1/48", gwCIDRSecond: "2001:db8:1:2::1/64"}, {gwCIDRFirst: "2001:db8:1:2::1/64", gwCIDRSecond: "fd00:1234::1/64"}}
		for _, tc := range testCases {
			gwIP, gwSubnet, err := net.ParseCIDR(tc.gwCIDRFirst)
			Expect(err).NotTo(HaveOccurred())
			gwnFirst := net.IPNet{IP: gwIP, Mask: gwSubnet.Mask}
			gwIP, gwSubnet, err = net.ParseCIDR(tc.gwCIDRSecond)
			Expect(err).NotTo(HaveOccurred())
			gwnSecond := net.IPNet{IP: gwIP, Mask: gwSubnet.Mask}
			var family, expNumAddrs int
			switch {
			case gwIP.To4() != nil:
				family = netlink.FAMILY_V4
				expNumAddrs = 1
			default:
				family = netlink.FAMILY_V6
				expNumAddrs = 2
			}
			subnetsOverlap := gwnFirst.Contains(gwnSecond.IP) || gwnSecond.Contains(gwnFirst.IP)
			err = originalNS.Do(func(ns.NetNS) error {
				defer GinkgoRecover()
				bridge, _, err := setupBridge(conf)
				Expect(err).NotTo(HaveOccurred())
				checkBridgeIPs := func(cidr0, cidr1 string) {
					addrs, err := netlink.AddrList(bridge, family)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(addrs)).To(Equal(expNumAddrs))
					addr := addrs[0].IPNet.String()
					Expect(addr).To(Equal(cidr0))
					if cidr1 != "" {
						addr = addrs[1].IPNet.String()
						Expect(addr).To(Equal(cidr1))
					}
				}
				Expect(conf.ForceAddress).To(Equal(false))
				err = ensureBridgeAddr(bridge, family, &gwnFirst, conf.ForceAddress)
				Expect(err).NotTo(HaveOccurred())
				checkBridgeIPs(tc.gwCIDRFirst, "")
				err = ensureBridgeAddr(bridge, family, &gwnSecond, false)
				if family == netlink.FAMILY_V4 || subnetsOverlap {
					Expect(err).To(HaveOccurred())
					checkBridgeIPs(tc.gwCIDRFirst, "")
				} else {
					Expect(err).NotTo(HaveOccurred())
					expNumAddrs++
					checkBridgeIPs(tc.gwCIDRSecond, tc.gwCIDRFirst)
				}
				err = ensureBridgeAddr(bridge, family, &gwnSecond, true)
				Expect(err).NotTo(HaveOccurred())
				if family == netlink.FAMILY_V4 || subnetsOverlap {
					checkBridgeIPs(tc.gwCIDRSecond, "")
				} else {
					checkBridgeIPs(tc.gwCIDRSecond, tc.gwCIDRFirst)
				}
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			delBridgeAddrs(originalNS)
		}
	})
	It("ensure promiscuous mode on bridge", func() {
		const IFNAME = "bridge0"
		const EXPECTED_IP = "10.0.0.0/8"
		const CHANGED_EXPECTED_IP = "10.1.2.3/16"
		conf := &NetConf{NetConf: types.NetConf{CNIVersion: "0.3.1", Name: "testConfig", Type: "bridge"}, BrName: IFNAME, IsGW: true, IPMasq: false, HairpinMode: false, PromiscMode: true, MTU: 5000}
		err := originalNS.Do(func(ns.NetNS) error {
			defer GinkgoRecover()
			_, _, err := setupBridge(conf)
			Expect(err).NotTo(HaveOccurred())
			Expect(conf.ForceAddress).To(Equal(false))
			link, err := netlink.LinkByName("bridge0")
			Expect(err).NotTo(HaveOccurred())
			Expect(link.Attrs().Promisc).To(Equal(1))
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})
	It("creates a bridge with a stable MAC addresses", func() {
		testCases := []testCase{{subnet: "10.1.2.0/24"}, {subnet: "2001:db8:42::/64"}}
		for _, tc := range testCases {
			tc.cniVersion = "0.3.1"
			_, _, err := setupBridge(tc.netConf())
			link, err := netlink.LinkByName(BRNAME)
			Expect(err).NotTo(HaveOccurred())
			origMac := link.Attrs().HardwareAddr
			cmdAddDelTest(originalNS, tc)
			link, err = netlink.LinkByName(BRNAME)
			Expect(err).NotTo(HaveOccurred())
			Expect(link.Attrs().HardwareAddr).To(Equal(origMac))
		}
	})
})
