package hwaddr_test

import (
	"net"
	"github.com/containernetworking/plugins/pkg/utils/hwaddr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hwaddr", func() {
	Context("Generate Hardware Address", func() {
		It("generate hardware address based on ipv4 address", func() {
			testCases := []struct {
				ip		net.IP
				expectedMAC	net.HardwareAddr
			}{{ip: net.ParseIP("10.0.0.2"), expectedMAC: (net.HardwareAddr)(append(hwaddr.PrivateMACPrefix, 0x0a, 0x00, 0x00, 0x02))}, {ip: net.ParseIP("10.250.0.244"), expectedMAC: (net.HardwareAddr)(append(hwaddr.PrivateMACPrefix, 0x0a, 0xfa, 0x00, 0xf4))}, {ip: net.ParseIP("172.17.0.2"), expectedMAC: (net.HardwareAddr)(append(hwaddr.PrivateMACPrefix, 0xac, 0x11, 0x00, 0x02))}, {ip: net.IPv4(byte(172), byte(17), byte(0), byte(2)), expectedMAC: (net.HardwareAddr)(append(hwaddr.PrivateMACPrefix, 0xac, 0x11, 0x00, 0x02))}}
			for _, tc := range testCases {
				mac, err := hwaddr.GenerateHardwareAddr4(tc.ip, hwaddr.PrivateMACPrefix)
				Expect(err).NotTo(HaveOccurred())
				Expect(mac).To(Equal(tc.expectedMAC))
			}
		})
		It("return error if input is not ipv4 address", func() {
			testCases := []net.IP{net.ParseIP(""), net.ParseIP("2001:db8:0:1:1:1:1:1")}
			for _, tc := range testCases {
				_, err := hwaddr.GenerateHardwareAddr4(tc, hwaddr.PrivateMACPrefix)
				Expect(err).To(BeAssignableToTypeOf(hwaddr.SupportIp4OnlyErr{}))
			}
		})
		It("return error if prefix is invalid", func() {
			_, err := hwaddr.GenerateHardwareAddr4(net.ParseIP("10.0.0.2"), []byte{0x58})
			Expect(err).To(BeAssignableToTypeOf(hwaddr.InvalidPrefixLengthErr{}))
		})
	})
})
