package main

import (
	"commons"
	"commons/as"
	"discovery"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"lua_binding"
	"metrics"
	"os"
	"snmp"
	"strings"
	"time"
	"web"
)

var (
	depth       = flag.Int("depth", 5, "the depth")
	timeout     = flag.Int("timeout", 5, "the timeout")
	network     = flag.String("ip-range", "", "the ip range")
	communities = flag.String("communities", "public", "the community")

	address   = flag.String("http", ":7070", "the address of http")
	directory = flag.String("directory", ".", "the static directory of http")
	cookies   = flag.String("cookies", "", "the static directory of http")
)

func mainHandle(rw *web.Context) {
	errFile := "_log_/error.html"
	_, err := os.Stat(errFile)
	if err == nil || os.IsExist(err) {
		content, _ := ioutil.ReadFile(errFile)
		rw.WriteString(string(content))
		return
	}
	rw.WriteString("Hello, World!")
}

func main() {

	flag.Parse()
	targets := flag.Args()
	if nil == targets || 0 != len(targets) {
		flag.Usage()
		return
	}

	svr := web.NewServer()
	svr.Config.Name = "meijing-bridge v1.0"
	svr.Config.Address = *address
	svr.Config.StaticDirectory = *directory
	svr.Config.CookieSecret = *cookies
	svr.Get("/", mainHandle)

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
	snmp_drv := snmp.NewSnmpDriver(time.Duration(*timeout)*time.Second, drvMgr)
	if err = snmp_drv.Start(); nil != err {
		fmt.Println(err)
		return
	}

	drvMgr.Register("snmp", snmp_drv)

	lua_drv := lua_binding.NewLuaDriver(time.Duration(*timeout)*time.Second, drvMgr)
	if err = lua_drv.Start(); nil != err {
		fmt.Println(err)
		return
	}
	drvMgr.Register("lua", lua_drv)

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

	go svr.Run()
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
