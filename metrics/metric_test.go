package metrics

import (
	"log"
	"lua_binding"
	"testing"
	"time"
)

func TestRoutes(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := lua_binding.NewLuaDriver(1*time.Second, nil)
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestRoutes"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "metric_tests", "target": "unit_test"}
	v := drv.Get(params)
	if v.HasError() {
		t.Error(v.Error())
		return
	}

	s, ok := v.Value().AsString()
	if nil != ok {
		t.Errorf("return is not a string, %T", v)
		return
	}

	if "ok" != s {
		t.Errorf("return != 'ok', it is %s", s)
		return
	}
}
