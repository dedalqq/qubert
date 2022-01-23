package interfaces

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/client4"
	"github.com/vishvananda/netlink"

	"qubert/logger"
)

type dhcpClientOptions struct {
	dev        string
	addr       *net.IPNet
	wg         sync.WaitGroup
	cancelFunc func()
}

func (o *dhcpClientOptions) stop() {
	o.cancelFunc()
	o.wg.Wait()
}

type dhcpClientManager struct {
	m sync.Mutex

	ctx    context.Context
	log    *logger.Logger
	update func()

	options []*dhcpClientOptions
}

func NewDHCPClientManager(ctx context.Context, log *logger.Logger, update func()) *dhcpClientManager {
	return &dhcpClientManager{
		ctx:    ctx,
		log:    log,
		update: update,
	}
}

func (c *dhcpClientManager) getDHCPOptions(dev string) *dhcpClientOptions {
	c.m.Lock()
	defer c.m.Unlock()

	for _, o := range c.options {
		if o.dev == dev {
			return o
		}
	}

	return nil
}

func (c *dhcpClientManager) runDHCPClient(link netlink.Link) {
	c.m.Lock()
	defer c.m.Unlock()

	for _, o := range c.options {
		if o.dev == link.Attrs().Name {
			return
		}
	}

	ctx, cancelFunc := context.WithCancel(c.ctx)

	opt := dhcpClientOptions{
		dev:        link.Attrs().Name,
		cancelFunc: cancelFunc,
	}

	c.options = append(c.options, &opt)

	opt.wg.Add(1)
	go func() {
		defer cancelFunc()
		defer opt.wg.Done()
		defer func() {
			c.m.Lock()
			defer c.m.Unlock()

			for i, o := range c.options {
				if o == &opt {
					c.options = append(c.options[:i], c.options[i+1:]...)

					return
				}
			}
		}()

		cl := client4.NewClient()
		t := time.NewTimer(0)

	mainLoop:
		for {

			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}

			conversation, err := cl.Exchange(link.Attrs().Name)
			if err != nil {
				//log.Error(err)
			}

			for _, cv := range conversation {
				if cv.MessageType() != dhcpv4.MessageTypeAck {
					continue
				}

				newIP := net.IPNet{
					IP:   cv.YourIPAddr,
					Mask: cv.SubnetMask(),
				}

				addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
				if err != nil {
					//log.Error(err)
				}

				if opt.addr != nil && opt.addr.String() != newIP.String() {
					for _, a := range addrs {
						if a.String() == opt.addr.String() {
							err = netlink.AddrDel(link, &netlink.Addr{IPNet: opt.addr})
							if err != nil {
								//log.Error(err)
							}

							break
						}
					}
				}

				opt.addr = &newIP

				for _, a := range addrs {
					if newIP.String() == a.String() {
						t = time.NewTimer(cv.IPAddressLeaseTime(time.Minute) / 2)
						c.update()
						continue mainLoop
					}
				}

				err = netlink.AddrAdd(link, &netlink.Addr{IPNet: &newIP})
				if err != nil {
					//log.Error(err)
				}

				t = time.NewTimer(cv.IPAddressLeaseTime(time.Minute) / 2)
				c.update()
				continue mainLoop

				// TODO routes and DNS
			}

			t = time.NewTimer(time.Second * 5)
			continue mainLoop
		}
	}()
}
