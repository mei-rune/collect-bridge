package main

import (
	"fmt"
	"license"
)

func main() {
	bs, code, e := license.Get("https://127.0.0.1:10443")
	if nil != e {
		fmt.Println(e)
		return
	}
	fmt.Println(code)
	fmt.Println(string(bs))
}
