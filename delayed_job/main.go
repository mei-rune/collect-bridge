package main

import (
	"commons"
	"flag"
	"fmt"
	"github.com/runner-mei/delayed_job"
)

func main() {
	flag.Parse()
	if nil != flag.Args() && 0 != len(flag.Args()) {
		flag.Usage()
		return
	}
	if e := commons.LoadConfig(nil); nil != e {
		return e
	}

	e := delayed_job.Main()
	if nil != e {
		fmt.Println(e)
		return
	}
}
