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
	go func() {
		for i := 1; ; i++ {
			err = icmp.Send(targets[0], nil)
			if nil != err {
				fmt.Println(err)
				break
			}

			if i >= 5 {
				break
			}
			time.Second(5 * time.Millisecond)
		}
	}()

	for {
		ra, _, err := icmp.Recv(time.Second)
		if nil != err {
			return
		}
		fmt.Println(ra)
	}

}
