package main

import (
	"commons"
	"flag"
	"fmt"
	"net"
	"snmp"
	"sync/atomic"
	"time"
)

var (
	laddr   = flag.String("laddr", "0.0.0.0:0", "the address of bind, default: '0.0.0.0:0'")
	network = flag.String("network", "udp4", "the family of address, default: 'udp4'")
	timeout = flag.Int("timeout", 5, "the family of address, default: '5'")
	port    = flag.String("port", "161", "the port of address, default: '161'")
)

func main() {
	flag.Parse()

	targets := flag.Args()
	if nil == targets || 1 != len(targets) {
		flag.Usage()
		return
	}

	scanner, err := snmp.NewPinger(*network, *laddr)
	if nil != err {
		fmt.Println(err)
		return
	}
	defer scanner.Close()

	ip_range, err := commons.ParseIPRange(targets[0])
	if nil != err {
		fmt.Println(err)
		return
	}
	var is_stopped int32 = 0
	go func() {
		for _, v := range []snmp.SnmpVersion{snmp.SNMP_V2C, snmp.SNMP_V3} {
			ip_range.Reset()
			for ip_range.HasNext() {
				err = scanner.Send(net.JoinHostPort(ip_range.Current().String(), *port), v)
				if nil != err {
					fmt.Println(err)
					goto end
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
