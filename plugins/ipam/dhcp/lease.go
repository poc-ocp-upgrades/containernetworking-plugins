package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"github.com/d2g/dhcp4"
	"github.com/d2g/dhcp4client"
	"github.com/vishvananda/netlink"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/plugins/pkg/ns"
)

const resendDelay0 = 4 * time.Second
const resendDelayMax = 32 * time.Second
const (
	leaseStateBound	= iota
	leaseStateRenewing
	leaseStateRebinding
)

type DHCPLease struct {
	clientID	string
	ack		*dhcp4.Packet
	opts		dhcp4.Options
	link		netlink.Link
	renewalTime	time.Time
	rebindingTime	time.Time
	expireTime	time.Time
	stopping	uint32
	stop		chan struct{}
	wg		sync.WaitGroup
}

func AcquireLease(clientID, netns, ifName string) (*DHCPLease, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	errCh := make(chan error, 1)
	l := &DHCPLease{clientID: clientID, stop: make(chan struct{})}
	log.Printf("%v: acquiring lease", clientID)
	l.wg.Add(1)
	go func() {
		errCh <- ns.WithNetNSPath(netns, func(_ ns.NetNS) error {
			defer l.wg.Done()
			link, err := netlink.LinkByName(ifName)
			if err != nil {
				return fmt.Errorf("error looking up %q: %v", ifName, err)
			}
			l.link = link
			if err = l.acquire(); err != nil {
				return err
			}
			log.Printf("%v: lease acquired, expiration is %v", l.clientID, l.expireTime)
			errCh <- nil
			l.maintain()
			return nil
		})
	}()
	if err := <-errCh; err != nil {
		return nil, err
	}
	return l, nil
}
func (l *DHCPLease) Stop() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if atomic.CompareAndSwapUint32(&l.stopping, 0, 1) {
		close(l.stop)
	}
	l.wg.Wait()
}
func (l *DHCPLease) acquire() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	c, err := newDHCPClient(l.link)
	if err != nil {
		return err
	}
	defer c.Close()
	if (l.link.Attrs().Flags & net.FlagUp) != net.FlagUp {
		log.Printf("Link %q down. Attempting to set up", l.link.Attrs().Name)
		if err = netlink.LinkSetUp(l.link); err != nil {
			return err
		}
	}
	pkt, err := backoffRetry(func() (*dhcp4.Packet, error) {
		ok, ack, err := c.Request()
		switch {
		case err != nil:
			return nil, err
		case !ok:
			return nil, fmt.Errorf("DHCP server NACK'd own offer")
		default:
			return &ack, nil
		}
	})
	if err != nil {
		return err
	}
	return l.commit(pkt)
}
func (l *DHCPLease) commit(ack *dhcp4.Packet) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	opts := ack.ParseOptions()
	leaseTime, err := parseLeaseTime(opts)
	if err != nil {
		return err
	}
	rebindingTime, err := parseRebindingTime(opts)
	if err != nil || rebindingTime > leaseTime {
		rebindingTime = leaseTime * 85 / 100
	}
	renewalTime, err := parseRenewalTime(opts)
	if err != nil || renewalTime > rebindingTime {
		renewalTime = leaseTime / 2
	}
	now := time.Now()
	l.expireTime = now.Add(leaseTime)
	l.renewalTime = now.Add(renewalTime)
	l.rebindingTime = now.Add(rebindingTime)
	l.ack = ack
	l.opts = opts
	return nil
}
func (l *DHCPLease) maintain() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	state := leaseStateBound
	for {
		var sleepDur time.Duration
		switch state {
		case leaseStateBound:
			sleepDur = l.renewalTime.Sub(time.Now())
			if sleepDur <= 0 {
				log.Printf("%v: renewing lease", l.clientID)
				state = leaseStateRenewing
				continue
			}
		case leaseStateRenewing:
			if err := l.renew(); err != nil {
				log.Printf("%v: %v", l.clientID, err)
				if time.Now().After(l.rebindingTime) {
					log.Printf("%v: renawal time expired, rebinding", l.clientID)
					state = leaseStateRebinding
				}
			} else {
				log.Printf("%v: lease renewed, expiration is %v", l.clientID, l.expireTime)
				state = leaseStateBound
			}
		case leaseStateRebinding:
			if err := l.acquire(); err != nil {
				log.Printf("%v: %v", l.clientID, err)
				if time.Now().After(l.expireTime) {
					log.Printf("%v: lease expired, bringing interface DOWN", l.clientID)
					l.downIface()
					return
				}
			} else {
				log.Printf("%v: lease rebound, expiration is %v", l.clientID, l.expireTime)
				state = leaseStateBound
			}
		}
		select {
		case <-time.After(sleepDur):
		case <-l.stop:
			if err := l.release(); err != nil {
				log.Printf("%v: failed to release DHCP lease: %v", l.clientID, err)
			}
			return
		}
	}
}
func (l *DHCPLease) downIface() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if err := netlink.LinkSetDown(l.link); err != nil {
		log.Printf("%v: failed to bring %v interface DOWN: %v", l.clientID, l.link.Attrs().Name, err)
	}
}
func (l *DHCPLease) renew() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	c, err := newDHCPClient(l.link)
	if err != nil {
		return err
	}
	defer c.Close()
	pkt, err := backoffRetry(func() (*dhcp4.Packet, error) {
		ok, ack, err := c.Renew(*l.ack)
		switch {
		case err != nil:
			return nil, err
		case !ok:
			return nil, fmt.Errorf("DHCP server did not renew lease")
		default:
			return &ack, nil
		}
	})
	if err != nil {
		return err
	}
	l.commit(pkt)
	return nil
}
func (l *DHCPLease) release() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	log.Printf("%v: releasing lease", l.clientID)
	c, err := newDHCPClient(l.link)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Release(*l.ack); err != nil {
		return fmt.Errorf("failed to send DHCPRELEASE")
	}
	return nil
}
func (l *DHCPLease) IPNet() (*net.IPNet, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	mask := parseSubnetMask(l.opts)
	if mask == nil {
		return nil, fmt.Errorf("DHCP option Subnet Mask not found in DHCPACK")
	}
	return &net.IPNet{IP: l.ack.YIAddr(), Mask: mask}, nil
}
func (l *DHCPLease) Gateway() net.IP {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return parseRouter(l.opts)
}
func (l *DHCPLease) Routes() []*types.Route {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	routes := []*types.Route{}
	opt121_routes := parseCIDRRoutes(l.opts)
	if len(opt121_routes) > 0 {
		return append(routes, opt121_routes...)
	}
	routes = append(routes, parseRoutes(l.opts)...)
	if gw := l.Gateway(); gw != nil {
		_, defaultRoute, _ := net.ParseCIDR("0.0.0.0/0")
		routes = append(routes, &types.Route{Dst: *defaultRoute, GW: gw})
	}
	return routes
}
func jitter(span time.Duration) time.Duration {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return time.Duration(float64(span) * (2.0*rand.Float64() - 1.0))
}
func backoffRetry(f func() (*dhcp4.Packet, error)) (*dhcp4.Packet, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var baseDelay time.Duration = resendDelay0
	for i := 0; i < resendCount; i++ {
		pkt, err := f()
		if err == nil {
			return pkt, nil
		}
		log.Print(err)
		time.Sleep(baseDelay + jitter(time.Second))
		if baseDelay < resendDelayMax {
			baseDelay *= 2
		}
	}
	return nil, errNoMoreTries
}
func newDHCPClient(link netlink.Link) (*dhcp4client.Client, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	pktsock, err := dhcp4client.NewPacketSock(link.Attrs().Index)
	if err != nil {
		return nil, err
	}
	return dhcp4client.New(dhcp4client.HardwareAddr(link.Attrs().HardwareAddr), dhcp4client.Timeout(5*time.Second), dhcp4client.Broadcast(false), dhcp4client.Connection(pktsock))
}
