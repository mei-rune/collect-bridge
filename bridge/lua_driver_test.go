package main

import (
	"fmt"
	"log"
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
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver("")
	drv.Start()
	drv.Stop()
}

func checkReturn(t *testing.T, excepted, old, actual interface{}, err error, msg string) {
	if nil != err {
		t.Errorf(msg+err.Error()+" - %v", old)
	} else if excepted != actual {
		t.Errorf(msg+" %v !=  %v - %v", excepted, actual, old)
	}
}

func TestPushAny(t *testing.T) {

	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver("")
	drv.Name = "TestPushAny"
	drv.init_path = "lua_init_test_pushAny.lua"
	drv.Start()

	var old, res interface{}
	var err error

	old = pushAnyTest(drv, nil)
	checkReturn(t, nil, old, old, nil, "test nil - ")

	old = pushAnyTest(drv, true)
	if nil == old {
		panic("nil == old")
	}
	res, err = asBool(old)
	checkReturn(t, true, old, res, err, "test bool - ")

	old = pushAnyTest(drv, int8(8))
	res, err = asInt8(old)
	checkReturn(t, int8(8), old, res, err, "test int8 - ")

	old = pushAnyTest(drv, int16(16))
	res, err = asInt16(old)
	checkReturn(t, int16(16), old, res, err, "test int16 - ")

	old = pushAnyTest(drv, int32(32))
	res, err = asInt32(old)
	checkReturn(t, int32(32), old, res, err, "test int32 - ")

	old = pushAnyTest(drv, int64(64))
	res, err = asInt64(old)
	checkReturn(t, int64(64), old, res, err, "test int64 - ")

	old = pushAnyTest(drv, uint8(98))
	res, err = asUint8(old)
	checkReturn(t, uint8(98), old, res, err, "test uint8 - ")

	old = pushAnyTest(drv, uint16(916))
	res, err = asUint16(old)
	checkReturn(t, uint16(916), old, res, err, "test uint16 - ")

	old = pushAnyTest(drv, uint32(932))
	res, err = asUint32(old)
	checkReturn(t, uint32(932), old, res, err, "test uint32 - ")

	old = pushAnyTest(drv, uint64(964))
	res, err = asUint64(old)
	checkReturn(t, uint64(964), old, res, err, "test uint64 - ")

	old = pushAnyTest(drv, uint(123))
	res, err = asUint(old)
	checkReturn(t, uint(123), old, res, err, "test uint - ")

	old = pushAnyTest(drv, int(223))
	res, err = asInt(old)
	checkReturn(t, int(223), old, res, err, "test int - ")

	old = pushAnyTest(drv, "aa")
	res, err = asString(old)
	checkReturn(t, "aa", old, res, err, "test string - ")

	drv.Stop()
}

func TestInvoke(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver("")
	drv.Name = "test_invoke"
	drv.init_path = "lua_init_test.lua"
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
	log.SetFlags(log.Flags() | log.Lshortfile)

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
	log.SetFlags(log.Flags() | log.Lshortfile)

	doFunc(false, t)
}
