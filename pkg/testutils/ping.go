package testutils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"syscall"
)

func Ping(saddr, daddr string, isV6 bool, timeoutSec int) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	args := []string{"-c", "1", "-W", strconv.Itoa(timeoutSec), "-I", saddr, daddr}
	bin := "ping"
	if isV6 {
		bin = "ping6"
	}
	cmd := exec.Command(bin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			return fmt.Errorf("%v exit status %d: %s", args, e.Sys().(syscall.WaitStatus).ExitStatus(), stderr.String())
		default:
			return err
		}
	}
	return nil
}
