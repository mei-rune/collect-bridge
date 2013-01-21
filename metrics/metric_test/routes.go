package main

import (
	"io/ioutil"
	"log"
	"lua_binding"
	"os"
	"web"
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
	log.SetFlags(log.Flags() | log.Lshortfile)

	svr := web.NewServer()
	svr.Config.Name = "meijing-routes-test v1.0"
	svr.Config.Address = ":7070"
	svr.Get("/", mainHandle)

	go svr.Run()

	drv := lua_binding.NewLuaDriver()
	drv.Name = "TestRoutes"
	drv.Start()

	defer func() {
		drv.Stop()
	}()

	params := map[string]string{"schema": "route_tests", "target": "unit_test"}
	v, e := drv.Get(params)
	if nil != e {
		log.Println(e)
		return
	}

	s, ok := v.(string)
	if !ok {
		log.Println(v, e)
		log.Printf("return is not a string, %T", v)
		return
	}

	if "ok" != s {
		log.Println(v, e)
		log.Printf("return != 'ok', it is %s", s)
		return
	}
}
