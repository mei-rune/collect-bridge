package metrics

import (
	"log"
	"lua_binding"
	"testing"
)

func TestRoutes(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := lua_binding.NewLuaDriver()
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestRoutes"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "metric_tests", "target": "unit_test"}
	v, e := drv.Get(params)
	if nil != e {
		t.Error(e)
		return
	}

	s, ok := v.(string)
	if !ok {
		t.Log(v, e)
		t.Errorf("return is not a string, %T", v)
		return
	}

	if "ok" != s {
		t.Log(v, e)
		t.Errorf("return != 'ok', it is %s", s)
		return
	}
}
