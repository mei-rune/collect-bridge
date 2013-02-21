package main

import (
	"commons"
	"commons/netutils"
	"flag"
	"fmt"
	"net"
	"snmp"
	"strings"
	"sync/atomic"
	"time"
)

var (
	laddr       = flag.String("laddr", "0.0.0.0:0", "the address of bind, default: '0.0.0.0:0'")
	network     = flag.String("network", "udp4", "the family of address, default: 'udp4'")
	timeout     = flag.Int("timeout", 5, "the family of address, default: '5'")
	port        = flag.String("port", "161", "the port of address, default: '161'")
	communities = flag.String("communities", "public;public1", "the community of snmp")
)

func main() {
	flag.Parse()

	targets := flag.Args()
	if nil == targets || 1 != len(targets) {
		flag.Usage()
		return
	}

	scanner, err := snmp.NewPinger(*network, *laddr, 256)
	if nil != err {
		fmt.Println(err)
		return
	}
	defer scanner.Close()

	ip_range, err := netutils.ParseIPRange(targets[0])
	if nil != err {
		fmt.Println(err)
		return
	}
	var is_stopped int32 = 0
	go func() {
		for _, v := range []snmp.SnmpVersion{snmp.SNMP_V2C, snmp.SNMP_V3} {
			ip_range.Reset()

			for j, community := range strings.Split(*communities, ";") {
				if j != 0 {
					time.Sleep(500 * time.Millisecond)
				}

				for ip_range.HasNext() {
					err = scanner.Send(net.JoinHostPort(ip_range.Current().String(), *port), v, community)
					if nil != err {
						fmt.Println(err)
						goto end
					}
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	end:
		atomic.StoreInt32(&is_stopped, 1)
	}()

	for {
		ra, t, err := scanner.Recv(time.Duration(*timeout) * time.Second)
		if nil != err {
			if !commons.IsTimeout(err) {
				fmt.Println(err)
			} else if 0 == atomic.LoadInt32(&is_stopped) {
				continue
			}
			return
		}
		fmt.Println(ra, t)
	}

}
