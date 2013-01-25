package lua_binding

import (
	"log"
	"strings"
	"testing"
	"time"
)

func TestArguments(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestArguments"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "arguments_test", "target": "unit_test"}
	v, e := drv.Get(params)
	if nil != e {
		t.Error(e)
		return
	}

	s, ok := v["value"].(string)
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

func TestExistFile(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestExistFile"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_exist_file", "file": "a.txt"}
	v, e := drv.Get(params)
	if nil != e {
		t.Error(e)
		t.FailNow()
	}

	s, ok := v["value"].(bool)
	if !ok {
		t.Log(v, e)
		t.Errorf("return is not a bool, %T", v)
		return
	}

	if !s {
		t.Log(v, e)
		t.Errorf("return != 'true', it is %s", s)
		return
	}

	params = map[string]string{"schema": "test_exist_file", "file": "aaaaaaaaaaaaaaaaa.txt"}
	v, e = drv.Get(params)
	if nil != e {
		t.Error(e)
		t.FailNow()
	}

	s, ok = v["value"].(bool)
	if !ok {
		t.Log(v, e)
		t.Errorf("return is not a bool, %T", v)
		return
	}

	if s {
		t.Log(v, e)
		t.Errorf("return != 'false', it is %s", s)
		return
	}
}

func TestExistDirectory(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestExistDirectory"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_exist_directory", "path": "enumerate_files"}
	v, e := drv.Get(params)
	if nil != e {
		t.Error(e)
		t.FailNow()
	}

	s, ok := v["value"].(bool)
	if !ok {
		t.Log(v, e)
		t.Errorf("return is not a bool, %T", v)
		return
	}

	if !s {
		t.Log(v, e)
		t.Errorf("return != 'true', it is %s", s)
		return
	}

	params = map[string]string{"schema": "test_exist_directory", "path": "aaaaaaaaaaaaaaaaa"}
	v, e = drv.Get(params)
	if nil != e {
		t.Error(e)
		t.FailNow()
	}

	s, ok = v["value"].(bool)
	if !ok {
		t.Log(v, e)
		t.Errorf("return is not a bool, %T", v)
		return
	}

	if s {
		t.Log(v, e)
		t.Errorf("return != 'false', it is %s", s)
		return
	}
}

func TestCleanPath(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestCleanPath"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_clean_path", "path": "/../a/b/../././/c"}
	v, e := drv.Get(params)
	if nil != e {
		t.Error(e)
		t.FailNow()
	}

	s, ok := v["value"].(string)
	if !ok {
		t.Log(v, e)
		t.Errorf("return is not a string, %T", v)
		return
	}

	if "/a/c" != s {
		t.Log(v, e)
		t.Errorf("return != 'true', it is %s", s)
		return
	}
}

func TestEnumerateFiles(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
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

	ss, ok := v["value"].([]interface{})
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
