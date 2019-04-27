package main_test

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"github.com/containernetworking/plugins/pkg/ns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Loopback", func() {
	var (
		networkNS	ns.NetNS
		containerID	string
		command		*exec.Cmd
		environ		[]string
	)
	BeforeEach(func() {
		command = exec.Command(pathToLoPlugin)
		var err error
		networkNS, err = ns.NewNS()
		Expect(err).NotTo(HaveOccurred())
		environ = []string{fmt.Sprintf("CNI_CONTAINERID=%s", containerID), fmt.Sprintf("CNI_NETNS=%s", networkNS.Path()), fmt.Sprintf("CNI_IFNAME=%s", "this is ignored"), fmt.Sprintf("CNI_ARGS=%s", "none"), fmt.Sprintf("CNI_PATH=%s", "/some/test/path")}
		command.Stdin = strings.NewReader(`{ "cniVersion": "0.1.0" }`)
	})
	AfterEach(func() {
		Expect(networkNS.Close()).To(Succeed())
	})
	Context("when given a network namespace", func() {
		It("sets the lo device to UP", func() {
			command.Env = append(environ, fmt.Sprintf("CNI_COMMAND=%s", "ADD"))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say(`{.*}`))
			Eventually(session).Should(gexec.Exit(0))
			var lo *net.Interface
			err = networkNS.Do(func(ns.NetNS) error {
				var err error
				lo, err = net.InterfaceByName("lo")
				return err
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(lo.Flags & net.FlagUp).To(Equal(net.FlagUp))
		})
		It("sets the lo device to DOWN", func() {
			command.Env = append(environ, fmt.Sprintf("CNI_COMMAND=%s", "DEL"))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say(``))
			Eventually(session).Should(gexec.Exit(0))
			var lo *net.Interface
			err = networkNS.Do(func(ns.NetNS) error {
				var err error
				lo, err = net.InterfaceByName("lo")
				return err
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(lo.Flags & net.FlagUp).NotTo(Equal(net.FlagUp))
		})
	})
})
