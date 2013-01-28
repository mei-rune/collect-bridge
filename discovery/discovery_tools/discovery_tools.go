package main

import (
	"commons/as"
	"discovery"
	"encoding/json"
	"flag"
	"fmt"
	"lua_binding"
	"time"
)

var (
	timeout = flag.Int("timeout", 5, "the timeout")
)

func main() {

	flag.Parse()
	targets := flag.Args()
	if nil == targets || 0 != len(targets) {
		flag.Usage()
		return
	}

	drvMgr := commons.NewDriverManager()
	drvMgr.Register("snmp", snmp.NewSnmpDriver(time.Duration(*timeout)*time.Second, drvMgr))
	drvMgr.Register("lua", lua_binding.NewLuaDriver(time.Duration(*timeout)*time.Second, drvMgr))
	drv := NewDiscoveryDriver(time.Duration(*timeout)*time.Second, drvMgr)
	res, err := drv.Create(map[string]string{"body": "{}"})
	if nil != err {
		fmt.Println(err)
		return
	}
	id, e := as.AsString(res["id"])
	if nil != e {
		fmt.Println(e)
		return
	}

	for {
		res, err := drv.Get(map[string]string{"id": id, "dst": "message"})
		if nil != err {
			fmt.Println(err)
			return
		}

	}
	res, err = drv.Get(map[string]string{"id": id})
	if nil != err {
		fmt.Println(err)
		return
	}

	bytes, e := json.Marshal(res)
	if nil != e {
		fmt.Println(e)
		return
	}
	fmt.Println(string(bytes))

	ok, err = drv.Delete(map[string]string{"id": id})
	if nil != err {
		fmt.Println(err)
		return
	}
}
