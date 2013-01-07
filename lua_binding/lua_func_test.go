package lua_binding

import (
	"log"
	"strings"
	"testing"
)

func TestEnumerateFiles(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver()
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestEnumerateFiles"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_enumerate_files"}
	v, e := drv.Get(params)
	if nil != e {
		t.Error(e)
		t.FailNow()
	}

	ss, ok := v.([]interface{})
	if !ok {
		t.Log(v, e)
		t.Errorf("return is not a []string, %T", v)
		t.FailNow()
	}

	if 3 != len(ss) {
		t.Log(v, e)
		t.Error("len(return) != 3")
		t.FailNow()
	}

	c := 0

	for _, a := range ss {
		s := a.(string)
		switch {

		case strings.HasSuffix(s, "c.txt"):
			c++
		case strings.HasSuffix(s, "a.txt"):
			c++
		case strings.HasSuffix(s, "b.txt"):
			c++
		}
	}

	if 3 != c {
		t.Log(v, e)
		t.Error("len(return) != 3")
		t.FailNow()
	}
}
