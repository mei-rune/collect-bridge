package main

import (
	"commons"
	"commons/as"
	"discovery"
	"encoding/json"
	"flag"
	"fmt"
	"lua_binding"
	"metrics"
	"snmp"
	"strings"
	"time"
)

var (
	depth       = flag.Int("depth", 5, "the depth")
	timeout     = flag.Int("timeout", 5, "the timeout")
	network     = flag.String("ip-range", "", "the ip range")
	communities = flag.String("communities", "public", "the community")
)

func main() {

	flag.Parse()
	targets := flag.Args()
	if nil == targets || 0 != len(targets) {
		flag.Usage()
		return
	}
	params := map[string]interface{}{}

	communities2 := strings.Split(*communities, ";")
	if nil != communities2 && 0 != len(communities2) {
		params["communities"] = communities2
	}

	network2 := strings.Split(*network, ";")
	if nil != network2 && 0 != len(network2) {
		params["ip-range"] = network2
	}
	params["depth"] = *depth
	js, err := json.Marshal(params)
	if nil != err {
		fmt.Println(err)
		return
	}

	drvMgr := commons.NewDriverManager()
	drvMgr.Register("snmp", snmp.NewSnmpDriver(time.Duration(*timeout)*time.Second, drvMgr))
	drvMgr.Register("lua", lua_binding.NewLuaDriver(time.Duration(*timeout)*time.Second, drvMgr))
	ms, err := metrics.NewMetrics(map[string]string{}, drvMgr)
	if nil != err {
		fmt.Println(err)
		return
	}
	drvMgr.Register("metrics", ms)

	drv := discovery.NewDiscoveryDriver(time.Duration(*timeout)*time.Second, drvMgr)
	res, err := drv.Create(map[string]string{"body": string(js)})
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
		values := commons.GetReturn(res)
		if nil == values {
			continue
		}
		messages, ok := values.([]string)
		if ok {
			isEnd := false
			for _, msg := range messages {
				if msg == "end" {
					isEnd = true
				}
				fmt.Println(msg)
			}
			if isEnd {
				break
			}
		}
	}
	res, err = drv.Get(map[string]string{"id": id})
	if nil != err {
		fmt.Println(err)
		return
	}

	bytes, e := json.MarshalIndent(res, "", "  ")
	if nil != e {
		fmt.Println(e)
		return
	}
	fmt.Println(string(bytes))

	_, err = drv.Delete(map[string]string{"id": id})
	if nil != err {
		fmt.Println(err)
		return
	}
}
