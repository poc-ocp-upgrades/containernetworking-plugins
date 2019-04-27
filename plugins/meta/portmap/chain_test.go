package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/coreos/go-iptables/iptables"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const TABLE = "filter"

var _ = Describe("chain tests", func() {
	var testChain chain
	var ipt *iptables.IPTables
	var cleanup func()
	BeforeEach(func() {
		currNs, err := ns.GetCurrentNS()
		Expect(err).NotTo(HaveOccurred())
		testNs, err := ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		tlChainName := fmt.Sprintf("cni-test-%d", rand.Intn(10000000))
		chainName := fmt.Sprintf("cni-test-%d", rand.Intn(10000000))
		testChain = chain{table: TABLE, name: chainName, entryChains: []string{tlChainName}, entryRules: [][]string{{"-d", "203.0.113.1"}}, rules: [][]string{{"-m", "comment", "--comment", "test 1", "-j", "RETURN"}, {"-m", "comment", "--comment", "test 2", "-j", "RETURN"}}}
		ipt, err = iptables.NewWithProtocol(iptables.ProtocolIPv4)
		Expect(err).NotTo(HaveOccurred())
		runtime.LockOSThread()
		err = testNs.Set()
		Expect(err).NotTo(HaveOccurred())
		err = ipt.ClearChain(TABLE, tlChainName)
		if err != nil {
			currNs.Set()
			Expect(err).NotTo(HaveOccurred())
		}
		cleanup = func() {
			if ipt == nil {
				return
			}
			ipt.ClearChain(TABLE, testChain.name)
			ipt.ClearChain(TABLE, tlChainName)
			ipt.DeleteChain(TABLE, testChain.name)
			ipt.DeleteChain(TABLE, tlChainName)
			currNs.Set()
		}
	})
	It("creates and destroys a chain", func() {
		defer cleanup()
		tlChainName := testChain.entryChains[0]
		err := ipt.Append(TABLE, tlChainName, "-m", "comment", "--comment", "canary value", "-j", "ACCEPT")
		Expect(err).NotTo(HaveOccurred())
		err = testChain.setup(ipt)
		Expect(err).NotTo(HaveOccurred())
		ok := false
		chains, err := ipt.ListChains(TABLE)
		Expect(err).NotTo(HaveOccurred())
		for _, chain := range chains {
			if chain == testChain.name {
				ok = true
				break
			}
		}
		if !ok {
			Fail("Could not find created chain")
		}
		haveRules, err := ipt.List(TABLE, tlChainName)
		Expect(err).NotTo(HaveOccurred())
		Expect(haveRules).To(Equal([]string{"-N " + tlChainName, "-A " + tlChainName + " -d 203.0.113.1/32 -j " + testChain.name, "-A " + tlChainName + ` -m comment --comment "canary value" -j ACCEPT`}))
		haveRules, err = ipt.List(TABLE, testChain.name)
		Expect(err).NotTo(HaveOccurred())
		Expect(haveRules).To(Equal([]string{"-N " + testChain.name, "-A " + testChain.name + ` -m comment --comment "test 1" -j RETURN`, "-A " + testChain.name + ` -m comment --comment "test 2" -j RETURN`}))
		err = testChain.teardown(ipt)
		Expect(err).NotTo(HaveOccurred())
		tlRules, err := ipt.List(TABLE, tlChainName)
		Expect(err).NotTo(HaveOccurred())
		Expect(tlRules).To(Equal([]string{"-N " + tlChainName, "-A " + tlChainName + ` -m comment --comment "canary value" -j ACCEPT`}))
		chains, err = ipt.ListChains(TABLE)
		Expect(err).NotTo(HaveOccurred())
		for _, chain := range chains {
			if chain == testChain.name {
				Fail("chain was not deleted")
			}
		}
	})
	It("creates chains idempotently", func() {
		defer cleanup()
		err := testChain.setup(ipt)
		Expect(err).NotTo(HaveOccurred())
		err = testChain.setup(ipt)
		Expect(err).NotTo(HaveOccurred())
		rules, err := ipt.List(TABLE, testChain.name)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(rules)).To(Equal(3))
	})
	It("deletes chains idempotently", func() {
		defer cleanup()
		err := testChain.setup(ipt)
		Expect(err).NotTo(HaveOccurred())
		err = testChain.teardown(ipt)
		Expect(err).NotTo(HaveOccurred())
		chains, err := ipt.ListChains(TABLE)
		for _, chain := range chains {
			if chain == testChain.name {
				Fail("Chain was not deleted")
			}
		}
		err = testChain.teardown(ipt)
		Expect(err).NotTo(HaveOccurred())
		chains, err = ipt.ListChains(TABLE)
		for _, chain := range chains {
			if chain == testChain.name {
				Fail("Chain was not deleted")
			}
		}
	})
})
