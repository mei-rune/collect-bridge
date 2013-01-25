package main

import (
	"commons"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
	"web"
)

var (
	address   = flag.String("http", ":7070", "the address of http")
	directory = flag.String("directory", ".", "the static directory of http")
	cookies   = flag.String("cookies", "", "the static directory of http")
	timeout   = flag.Int("timeout", 5, "the timeout of http")
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
	svr := web.NewServer()
	svr.Config.Name = "meijing-bridge v1.0"
	svr.Config.Address = *address
	svr.Config.StaticDirectory = *directory
	svr.Config.CookieSecret = *cookies
	svr.Get("/", mainHandle)
	drvMgr := commons.NewDriverManager()

	e := registerSNMP(svr, time.Duration(*timeout)*time.Second, drvMgr)
	if nil == e {
		fmt.Println(e)
		return
	}
	e = registerLua(svr, time.Duration(*timeout)*time.Second, drvMgr)
	if nil == e {
		fmt.Println(e)
		return
	}
	registerDriverBridge(svr, drvMgr)

	svr.Run()
}
