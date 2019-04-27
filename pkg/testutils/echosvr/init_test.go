package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestEchosvr(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testutils Echosvr Suite")
}
