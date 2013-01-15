package main

import (
	"log"
	"lua_binding"
)

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)

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
