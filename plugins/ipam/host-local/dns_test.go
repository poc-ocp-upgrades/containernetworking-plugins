package main

import (
	"io/ioutil"
	"os"
	"github.com/containernetworking/cni/pkg/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parsing resolv.conf", func() {
	It("parses a simple resolv.conf file", func() {
		contents := `
		nameserver 192.0.2.0
		nameserver 192.0.2.1
		`
		dns, err := parse(contents)
		Expect(err).NotTo(HaveOccurred())
		Expect(*dns).Should(Equal(types.DNS{Nameservers: []string{"192.0.2.0", "192.0.2.1"}}))
	})
	It("ignores comments", func() {
		dns, err := parse(`
nameserver 192.0.2.0
;nameserver 192.0.2.1
`)
		Expect(err).NotTo(HaveOccurred())
		Expect(*dns).Should(Equal(types.DNS{Nameservers: []string{"192.0.2.0"}}))
	})
	It("parses all fields", func() {
		dns, err := parse(`
nameserver 192.0.2.0
nameserver 192.0.2.2
domain example.com
;nameserver comment
#nameserver comment
search example.net example.org
search example.gov
options one two three
options four
`)
		Expect(err).NotTo(HaveOccurred())
		Expect(*dns).Should(Equal(types.DNS{Nameservers: []string{"192.0.2.0", "192.0.2.2"}, Domain: "example.com", Search: []string{"example.net", "example.org", "example.gov"}, Options: []string{"one", "two", "three", "four"}}))
	})
})

func parse(contents string) (*types.DNS, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	f, err := ioutil.TempFile("", "host_local_resolv")
	defer f.Close()
	defer os.Remove(f.Name())
	if err != nil {
		return nil, err
	}
	if _, err := f.WriteString(contents); err != nil {
		return nil, err
	}
	return parseResolvConf(f.Name())
}
