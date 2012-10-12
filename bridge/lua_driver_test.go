package main

import (
	"fmt"
	"strings"
	"testing"
	"unsafe"
)

const (
	s1 = `
  function receive ()
	`
)

func TestSpawn(t *testing.T) {
	drv := NewLuaDriver("")
	drv.Start()
	drv.Stop()
}

func TestInvoke(t *testing.T) {
	drv := NewLuaDriver("")
	drv.Name = "test_invoke"
	drv.Start()

	fmt.Printf("aaaaaaaaaa %v\n", unsafe.Pointer(drv.ls))
	v, e := drv.Get(nil)
	if nil != e {
		t.Errorf("execute get failed, " + e.Error())
	} else if s, _ := asString(v); "test ok" != s {
		t.Errorf("execute get faile, excepted value is 'ok', actual is %v", v)
	} else {
		t.Log("execute get ok")
	}
	fmt.Printf("aaaaaaaaaa %v\n", unsafe.Pointer(drv.ls))
	drv.Stop()
}

func TestSpawnWithInitScript(t *testing.T) {
	drv := NewLuaDriver(s1)
	drv.init_path = "aa"
	err := drv.Start()
	if nil == err {
		t.Errorf("test start lua failed, except return a error, actual return success")
	} else if !strings.Contains(err.Error(), "(to close 'function' at line 2)") {
		t.Errorf("test start lua failed, excepted value contains '(to close 'function' at line 2)', actual value is " + err.Error())
	}
}

func doFunc(b bool, t *testing.T) {
	if b {
		defer func() {
			t.Error("it is faile")
		}()
	}
}

func TestDefer(t *testing.T) {
	doFunc(false, t)
}
