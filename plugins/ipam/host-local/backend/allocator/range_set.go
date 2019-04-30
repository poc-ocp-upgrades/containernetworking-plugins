package allocator

import (
	"fmt"
	"net"
	"strings"
)

func (s *RangeSet) Contains(addr net.IP) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	r, _ := s.RangeFor(addr)
	return r != nil
}
func (s *RangeSet) RangeFor(addr net.IP) (*Range, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if err := canonicalizeIP(&addr); err != nil {
		return nil, err
	}
	for _, r := range *s {
		if r.Contains(addr) {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("%s not in range set %s", addr.String(), s.String())
}
func (s *RangeSet) Overlaps(p1 *RangeSet) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, r := range *s {
		for _, r1 := range *p1 {
			if r.Overlaps(&r1) {
				return true
			}
		}
	}
	return false
}
func (s *RangeSet) Canonicalize() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(*s) == 0 {
		return fmt.Errorf("empty range set")
	}
	fam := 0
	for i, _ := range *s {
		if err := (*s)[i].Canonicalize(); err != nil {
			return err
		}
		if i == 0 {
			fam = len((*s)[i].RangeStart)
		} else {
			if fam != len((*s)[i].RangeStart) {
				return fmt.Errorf("mixed address families")
			}
		}
	}
	l := len(*s)
	for i, r1 := range (*s)[:l-1] {
		for _, r2 := range (*s)[i+1:] {
			if r1.Overlaps(&r2) {
				return fmt.Errorf("subnets %s and %s overlap", r1.String(), r2.String())
			}
		}
	}
	return nil
}
func (s *RangeSet) String() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	out := []string{}
	for _, r := range *s {
		out = append(out, r.String())
	}
	return strings.Join(out, ",")
}
