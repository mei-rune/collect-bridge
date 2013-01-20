package main

import (
	"flag"
	"fmt"
	"nmap"
)

var (
	target = flag.String("address", "127.0.0.1", "the ping address, default: 127.0.0.1")
)

func main() {
	icmp, err := nmap.NewICMP("ip4:icmp", "", []byte("gogogogogogogogo"))
	if nil != err {
		fmt.Println(err)
		return
	}
	err = icmp.Send(*target, nil)
	if nil != err {
		fmt.Println(err)
		return
	}
	ra, _, err := icmp.Recv()
	if nil != err {
		fmt.Println(err)
		return
	}

	fmt.Println(ra)
}
