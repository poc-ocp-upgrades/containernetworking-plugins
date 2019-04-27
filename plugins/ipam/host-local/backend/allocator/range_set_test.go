package allocator

import (
	"net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("range sets", func() {
	It("should detect set membership correctly", func() {
		p := RangeSet{Range{Subnet: mustSubnet("192.168.0.0/24")}, Range{Subnet: mustSubnet("172.16.1.0/24")}}
		err := p.Canonicalize()
		Expect(err).NotTo(HaveOccurred())
		Expect(p.Contains(net.IP{192, 168, 0, 55})).To(BeTrue())
		r, err := p.RangeFor(net.IP{192, 168, 0, 55})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).To(Equal(&p[0]))
		r, err = p.RangeFor(net.IP{192, 168, 99, 99})
		Expect(r).To(BeNil())
		Expect(err).To(MatchError("192.168.99.99 not in range set 192.168.0.1-192.168.0.254,172.16.1.1-172.16.1.254"))
	})
	It("should discover overlaps within a set", func() {
		p := RangeSet{{Subnet: mustSubnet("192.168.0.0/20")}, {Subnet: mustSubnet("192.168.2.0/24")}}
		err := p.Canonicalize()
		Expect(err).To(MatchError("subnets 192.168.0.1-192.168.15.254 and 192.168.2.1-192.168.2.254 overlap"))
	})
	It("should discover overlaps outside a set", func() {
		p1 := RangeSet{{Subnet: mustSubnet("192.168.0.0/20")}}
		p2 := RangeSet{{Subnet: mustSubnet("192.168.2.0/24")}}
		p1.Canonicalize()
		p2.Canonicalize()
		Expect(p1.Overlaps(&p2)).To(BeTrue())
		Expect(p2.Overlaps(&p1)).To(BeTrue())
	})
})
