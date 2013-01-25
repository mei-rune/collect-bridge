package main

import (
	"commons"
	"commons/netutils"
	"flag"
	"fmt"
	"sync/atomic"
	"time"
)

var (
	msg     = flag.String("msg", "gogogogo", "the body of icmp message, default: 'gogogogo'")
	laddr   = flag.String("laddr", "", "the address of bind, default: ''")
	network = flag.String("network", "ip4:icmp", "the family of address, default: 'ip4:icmp'")
)

func main() {
	flag.Parse()

	targets := flag.Args()
	if nil == targets || 1 != len(targets) {
		flag.Usage()
		return
	}

	icmp, err := netutils.NewPinger(*network, *laddr, []byte(*msg))
	if nil != err {
		fmt.Println(err)
		return
	}
	defer func() {
		fmt.Println("exit")
		icmp.Close()
	}()

	ip_range, err := netutils.ParseIPRange(targets[0])
	if nil != err {
		fmt.Println(err)
		return
	}

	var is_stopped int32 = 0
	go func() {
		for ip_range.HasNext() {
			err = icmp.Send(ip_range.Current().String(), nil)
			if nil != err {
				fmt.Println(err)
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
		atomic.StoreInt32(&is_stopped, 1)
	}()

	for {
		ra, _, err := icmp.Recv(time.Second)
		if nil != err {
			if !commons.IsTimeout(err) {
				fmt.Println(err)
			} else if 0 == atomic.LoadInt32(&is_stopped) {
				continue
			}
			return
		}
		fmt.Println(ra)
	}

}
