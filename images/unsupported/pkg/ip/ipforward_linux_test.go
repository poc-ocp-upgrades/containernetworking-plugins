package ip

import (
	"io/ioutil"
	"os"
	"time"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IpforwardLinux", func() {
	It("echo1 must not write the file if content is 1", func() {
		file, err := ioutil.TempFile(os.TempDir(), "containernetworking")
		defer os.Remove(file.Name())
		err = echo1(file.Name())
		Expect(err).NotTo(HaveOccurred())
		statBefore, err := file.Stat()
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(100 * time.Millisecond)
		err = echo1(file.Name())
		Expect(err).NotTo(HaveOccurred())
		statAfter, err := file.Stat()
		Expect(err).NotTo(HaveOccurred())
		Expect(statBefore.ModTime()).To(Equal(statAfter.ModTime()))
	})
})
