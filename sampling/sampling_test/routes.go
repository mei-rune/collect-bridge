package main

import (
	"log"
	"lua_binding"
	"time"
)

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := lua_binding.NewLuaDriver(1*time.Second, nil)
	drv.Name = "TestRoutes"
	drv.Start()

	defer func() {
		drv.Stop()
	}()

	params := map[string]string{"schema": "route_tests", "target": "unit_test"}
	res := drv.Get(params)
	if res.HasError() {
		log.Println(res.Error())
		return
	}

	s, ok := res.InterfaceValue().(string)
	if !ok {
		log.Println(res)
		log.Printf("return is not a string, %T", res.InterfaceValue())
		return
	}

	if "ok" != s {
		log.Println(res)
		log.Printf("return != 'ok', it is %s", s)
		return
	}
}
