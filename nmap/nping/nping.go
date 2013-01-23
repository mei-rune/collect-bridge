package main

import (
	"flag"
	"fmt"
	"nmap"
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

	icmp, err := nmap.NewICMP(*network, *laddr, []byte(*msg))
	if nil != err {
		fmt.Println(err)
		return
	}
	defer icmp.Close()

	fmt.Println("ping ", targets[0])

	for i := 1; i < 100; i++ {
		err = icmp.Send(fmt.Sprintf("20.0.8.%d", i), nil)
		if nil != err {
			fmt.Println(err)
			return
		}
	}
	for {
		ra, _, err := icmp.Recv(time.Second)
		if nil != err {
			return
		}
		fmt.Println(ra)
	}

}
