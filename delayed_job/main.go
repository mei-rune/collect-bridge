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
	if e := commons.LoadDefaultProperties("data.", "db_drv", "db_url", "redis", map[string]string{"redis.host": "127.0.0.1",
		"redis.port":  "36379",
		"db.type":     "postgresql",
		"db.address":  "127.0.0.1",
		"db.port":     "35432",
		"db.schema":   "tpt_data",
		"db.username": "tpt",
		"db.password": "extreme"}); nil != e {
		fmt.Println(e)
		return
	}

	e := delayed_job.Main()
	if nil != e {
		fmt.Println(e)
		return
	}
}
