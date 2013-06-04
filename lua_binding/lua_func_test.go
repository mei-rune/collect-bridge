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
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestArguments"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "arguments_test", "target": "unit_test"}
	v := drv.Get(params)
	if v.HasError() {
		t.Error(v.ErrorMessage())
		return
	}

	s, ok := v.Value().AsString()
	if nil != ok {
		t.Log(v)
		t.Errorf("return is not a string, %T", v.Value().AsInterface())
		return
	}

	if "ok" != s {
		t.Log(v)
		t.Errorf("return != 'ok', it is %s", s)
		return
	}
}

func TestExistFile(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestExistFile"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_exist_file", "file": "a.txt"}

	v := drv.Get(params)
	if v.HasError() {
		t.Error(v.ErrorMessage())
		return
	}

	s, ok := v.Value().AsBool()
	if nil != ok {
		t.Log(v)
		t.Errorf("return is not a bool, %T", v.Value().AsInterface())
		return
	}

	if !s {
		t.Log(v)
		t.Errorf("return != 'true', it is %s", s)
		return
	}

	params = map[string]string{"schema": "test_exist_file", "file": "aaaaaaaaaaaaaaaaa.txt"}
	v = drv.Get(params)
	if v.HasError() {
		t.Error(v.ErrorMessage())
		return
	}

	b, ok := v.Value().AsBool()
	if nil != ok {
		t.Log(v)
		t.Errorf("return is not a bool, %T", v.Value().AsInterface())
		return
	}

	if b {
		t.Log(v)
		t.Errorf("return != 'false', it is %s", b)
		return
	}
}

func TestExistDirectory(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestExistDirectory"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_exist_directory", "path": "enumerate_files"}
	v := drv.Get(params)
	if v.HasError() {
		t.Error(v.ErrorMessage())
		return
	}

	s, ok := v.Value().AsBool()
	if nil != ok {
		t.Log(v)
		t.Errorf("return is not a bool, %T", v.Value().AsInterface())
		return
	}

	if !s {
		t.Log(v)
		t.Errorf("return != 'true', it is %s", s)
		return
	}

	params = map[string]string{"schema": "test_exist_directory", "path": "aaaaaaaaaaaaaaaaa"}
	v = drv.Get(params)
	if v.HasError() {
		t.Error(v.ErrorMessage())
		return
	}

	s, ok = v.Value().AsBool()
	if nil != ok {
		t.Log(v)
		t.Errorf("return is not a bool, %T", v.Value().AsInterface())
		return
	}

	if s {
		t.Log(v)
		t.Errorf("return != 'false', it is %s", s)
		return
	}
}

func TestCleanPath(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestCleanPath"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_clean_path", "path": "/../a/b/../././/c"}
	v := drv.Get(params)
	if v.HasError() {
		t.Error(v.ErrorMessage())
		return
	}

	s, ok := v.Value().AsString()
	if nil != ok {
		t.Log(v)
		t.Errorf("return is not a string, %T", v.Value().AsInterface())
		return
	}

	if "/a/c" != s {
		t.Log(v)
		t.Errorf("return != 'true', it is %s", s)
		return
	}
}

func TestEnumerateFiles(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestEnumerateFiles"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_enumerate_files"}
	v := drv.Get(params)
	if v.HasError() {
		t.Error(v.ErrorMessage())
		return
	}

	ss, ok := v.Value().AsInterface().([]interface{})
	if !ok {
		t.Log(v)
		t.Errorf("return is not a []string, %T", v.Value().AsInterface())
		t.FailNow()
	}

	if 3 != len(ss) {
		t.Log(v)
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
		t.Log(v)
		t.Error("len(return) != 3")
		t.FailNow()
	}
}
