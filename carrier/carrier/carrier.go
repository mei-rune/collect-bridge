package main

import (
	"carrier"
	"fmt"
)

func main() {
	e := carrier.Main(false)
	if nil != e {
		fmt.Println(e)
	}
}
