package lua_binding

import (
	"commons"
	c "commons/as"
	"commons/errutils"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSpawn(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.Start()
	drv.Stop()
}
func assertExceptedEqualActual(t *testing.T, excepted, actual interface{}, msg string) {
	if excepted != actual {
		t.Errorf(msg+" %v !=  %v", excepted, actual)
	}
}

func assertNil(t *testing.T, value interface{}, msg string) {
	if nil != value {
		t.Errorf(msg)
	}
}

func assertFalse(t *testing.T, cond bool, msg string) {
	if cond {
		t.Errorf(msg)
	}
}

func assertTrue(t *testing.T, cond bool, msg string) {
	if !cond {
		t.Errorf(msg)
	}
}

func checkReturn(t *testing.T, excepted, old, actual interface{}, err error, msg string) {
	if nil != err {
		t.Errorf(msg+err.Error()+" - %v", old)
	} else if excepted != actual {
		t.Errorf(msg+" %v !=  %v - %v", excepted, actual, old)
	}
}

func checkArray(t *testing.T, old interface{}) {
	var res interface{}
	var err error

	array, err := c.AsArray(old)
	if nil != err {
		t.Errorf("test array - "+err.Error()+" - %v", old)
	} else {
		res, err = c.AsInt(array[0])
		checkReturn(t, int(1), array[0], res, err, "test int in array - ")

		res, err = c.AsUint(array[1])
		checkReturn(t, uint(2), array[1], res, err, "test uint in array - ")

		res, err = c.AsString(array[2])
		checkReturn(t, "s1", array[2], res, err, "test string in array - ")
	}
}

func checkMap(t *testing.T, old interface{}) {

	var res interface{}
	var err error

	assoc, err := c.AsMap(old)
	if nil != err {
		t.Errorf("test map - "+err.Error()+" - %v", old)
	} else {
		fmt.Print(assoc)

		res, err = c.AsInt(assoc["a1"])
		checkReturn(t, int(1), assoc["a1"], res, err, "test int in map - ")

		res, err = c.AsUint(assoc["a2"])
		checkReturn(t, uint(2), assoc["a2"], res, err, "test uint in map - ")

		res, err = c.AsString(assoc["a3"])
		checkReturn(t, "s3", assoc["a3"], res, err, "test string in array - ")
	}
}

func TestParams(t *testing.T) {

	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestParams"
	drv.init_path = "test/lua_init_test_pushAny.lua"
	drv.Start()

	if !drv.IsRunning() {
		t.Errorf("TestParams - start failed")
		return
	}

	pushString(drv.LS, "test")
	pushParams(drv.LS, map[string]string{"a": "sa", "b": "sb"})
	ResumeLuaFiber(drv, 2)
	params, _ := toParams(drv.LS, 2)

	assertExceptedEqualActual(t, "sa", params["a"], "test params - ")
	assertExceptedEqualActual(t, "sb", params["b"], "test params - ")

	pushString(drv.LS, "test")
	pushParams(drv.LS, map[string]string{})
	ResumeLuaFiber(drv, 2)
	params, _ = toParams(drv.LS, 2)

	assertExceptedEqualActual(t, int(0), len(params), "test params - ")

	pushString(drv.LS, "test")
	pushParams(drv.LS, nil)
	ResumeLuaFiber(drv, 2)
	params, _ = toParams(drv.LS, 2)

	// A nil map is equivalent to an empty map except that no elements may be added. 
	assertExceptedEqualActual(t, int(0), len(params), "test params - ")

	drv.Stop()
}

func TestPushAny(t *testing.T) {

	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestPushAny"
	drv.init_path = "test/lua_init_test_pushAny.lua"
	drv.Start()

	if !drv.IsRunning() {
		t.Errorf("TestPushAny - start failed")
		return
	}

	var old, res interface{}
	var err error

	old = pushAnyTest(drv, nil)
	checkReturn(t, nil, old, old, nil, "test nil - ")

	old = pushAnyTest(drv, true)
	if nil == old {
		panic("nil == old")
	}
	res, err = c.AsBool(old)
	checkReturn(t, true, old, res, err, "test bool - ")

	old = pushAnyTest(drv, int8(8))
	res, err = c.AsInt8(old)
	checkReturn(t, int8(8), old, res, err, "test int8 - ")

	old = pushAnyTest(drv, int16(16))
	res, err = c.AsInt16(old)
	checkReturn(t, int16(16), old, res, err, "test int16 - ")

	old = pushAnyTest(drv, int32(32))
	res, err = c.AsInt32(old)
	checkReturn(t, int32(32), old, res, err, "test int32 - ")

	old = pushAnyTest(drv, int64(64))
	res, err = c.AsInt64(old)
	checkReturn(t, int64(64), old, res, err, "test int64 - ")

	old = pushAnyTest(drv, uint8(98))
	res, err = c.AsUint8(old)
	checkReturn(t, uint8(98), old, res, err, "test uint8 - ")

	old = pushAnyTest(drv, uint16(916))
	res, err = c.AsUint16(old)
	checkReturn(t, uint16(916), old, res, err, "test uint16 - ")

	old = pushAnyTest(drv, uint32(932))
	res, err = c.AsUint32(old)
	checkReturn(t, uint32(932), old, res, err, "test uint32 - ")

	old = pushAnyTest(drv, uint64(964))
	res, err = c.AsUint64(old)
	checkReturn(t, uint64(964), old, res, err, "test uint64 - ")

	old = pushAnyTest(drv, uint(123))
	res, err = c.AsUint(old)
	checkReturn(t, uint(123), old, res, err, "test uint - ")

	old = pushAnyTest(drv, int(223))
	res, err = c.AsInt(old)
	checkReturn(t, int(223), old, res, err, "test int - ")

	old = pushAnyTest(drv, "aa")
	res, err = c.AsString(old)
	checkReturn(t, "aa", old, res, err, "test string - ")

	old = pushAnyTest(drv, []interface{}{int8(1), uint8(2), "s1"})
	checkArray(t, old)

	old = pushAnyTest(drv, map[string]interface{}{"a1": 1, "a2": uint(2), "a3": "s3"})
	checkMap(t, old)

	old = pushAnyTest(drv, []interface{}{int8(1), uint8(2), "s1", map[string]interface{}{"a1": 1, "a2": uint(2), "a3": "s3"}})
	checkArray(t, old)

	array, err := c.AsArray(old)
	if nil == err {
		t.Log("test map in array")
		checkMap(t, array[3])
		t.Log("test map in array is ok")
	}

	old = pushAnyTest(drv, map[string]interface{}{"a1": 1, "a2": uint(2), "a3": "s3", "a4": []interface{}{int8(1), uint8(2), "s1"}})
	checkMap(t, old)

	assoc, err := c.AsMap(old)
	if nil == err {
		t.Log("test array in map")
		checkArray(t, assoc["a4"])
		t.Log("test array in map is ok")
	}

	drv.Stop()
}

func TestInvoke(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "test_invoke"
	drv.init_path = "test/lua_init_test.lua"
	drv.Start()

	if !drv.IsRunning() {
		t.Errorf("TestInvoke - start failed")
		return
	}

	v, e := drv.Get(nil)
	if nil != e {
		t.Errorf("execute get failed, " + e.Error())
	} else if s, _ := c.AsString(v["value"]); "test ok" != s {
		t.Errorf("execute get failed, excepted value is 'ok', actual is %v", v)
	} else {
		t.Log("execute get ok")
	}
	drv.Stop()
}

func TestInvokeScript(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestInvokeScript"
	drv.Start()

	if !drv.IsRunning() {
		t.Errorf("TestInvokeScript - start failed")
		return
	}

	params := map[string]string{"schema": "script", "script": "return {value= action..' ok'}, nil"}
	v, e := drv.Get(params)
	if nil != e {
		t.Errorf("execute get failed, " + e.Error())
	} else if s, _ := c.AsString(v["value"]); "get ok" != s {
		t.Errorf("execute get failed, excepted value is 'ok', actual is %v", v)
	} else {
		t.Log("execute get ok")
	}
	drv.Stop()
}

func TestInvokeScriptFailed(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestInvokeScriptFailed"
	drv.Start()

	if !drv.IsRunning() {
		t.Errorf("TestInvokeScriptFailed - start failed")
		return
	}

	params := map[string]string{"schema": "script", "script": "aa"}
	_, e := drv.Get(params)
	if nil == e {
		t.Errorf("execute get failed, except return error, actual return ok")
	} else if !strings.Contains(e.Error(), "syntax error near <eof>") {
		t.Errorf("execute get failed, except error contains 'syntax error near <eof>', actual return - " + e.Error())
	}
	drv.Stop()
	drv.drvMgr.Unregister("test")
}

type TestDriver struct {
	create, get, put     interface{}
	get_error, put_error commons.RuntimeError

	delete                     bool
	create_error, delete_error commons.RuntimeError
}

func (bridge *TestDriver) Get(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return map[string]interface{}{"value": bridge.get}, bridge.get_error
}

func (bridge *TestDriver) Put(params map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return map[string]interface{}{"value": bridge.put}, bridge.put_error
}

func (bridge *TestDriver) Create(map[string]string) (map[string]interface{}, commons.RuntimeError) {
	return map[string]interface{}{"value": bridge.create}, bridge.create_error
}

func (bridge *TestDriver) Delete(map[string]string) (bool, commons.RuntimeError) {
	return bridge.delete, bridge.delete_error
}

func TestInvokeAndCallback(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestInvokeAndCallback"
	drv.Start()

	if !drv.IsRunning() {
		t.Errorf("TestInvokeAndCallback - start failed")
		return
	}

	td := &TestDriver{get: "get12", put: "put12", create: true, delete: true}
	drv.drvMgr.Register("test_dumy", td)

	defer func() {
		drv.Stop()
		drv.drvMgr.Unregister("test_dumy")
	}()

	params := map[string]string{"schema": "script", "script": "mj.log(mj.DEBUG, 'log a test log.')\nreturn mj.execute('test_dumy', action, params)"}
	v, e := drv.Get(params)
	if nil != e {
		t.Errorf("execute get failed, " + e.Error())
	} else if s, _ := c.AsString(v["value"]); "get12" != s {
		t.Errorf("execute get failed, excepted value is 'ok', actual is %v", v)
	} else {
		t.Log("execute get ok")
	}
}

func testResult(t *testing.T, drv *LuaDriver, excepted_value interface{}, excepted_error string, actual_value interface{}, actual_error commons.RuntimeError) {
	if nil == actual_error {
		if "" != excepted_error {
			t.Errorf("execute failed, excepted error is %v, actual error is nil", excepted_error)
		}
	} else if actual_error.Error() != excepted_error {
		t.Errorf("execute failed, excepted error is %v, actual error is %v", excepted_error, actual_error.Error())
	}

	if nil == actual_value {
		if nil != excepted_value {
			t.Errorf("execute failed, excepted value is %v, actual value is nil", excepted_value)
		}
	} else if !reflect.DeepEqual(actual_value, excepted_value) {
		t.Errorf("execute failed, excepted value is '%v', actual value is %v", excepted_value, actual_value)
	} else {
		t.Log("execute ok")
	}
}

func TestInvokeModule(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestInvokeModule"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_invoke_module"}
	v, e := drv.Get(params)
	testResult(t, drv, map[string]interface{}{"value": "get test ok test1whj23"}, "", v, e)
	v, e = drv.Put(params)
	testResult(t, drv, map[string]interface{}{"value": "put test ok test1whj23"}, "", v, e)
	v, e = drv.Create(params)
	testResult(t, drv, map[string]interface{}{"id": "2328"}, "create test ok test1whj23", v, e)
	b, e := drv.Delete(params)
	testResult(t, drv, false, "delete test ok test1whj23", b, e)
}

func TestInvokeModuleFailed(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestInvokeModuleFailed"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_invoke_module_failed"}
	v, e := drv.Get(params)
	if nil == v {
		testResult(t, drv, nil, "get error for test_invoke_module_failed", nil, e)
	} else {
		testResult(t, drv, nil, "get error for test_invoke_module_failed", v, e)
	}
	v, e = drv.Put(params)
	if nil == v {
		testResult(t, drv, nil, "put error for test_invoke_module_failed", nil, e)
	} else {
		testResult(t, drv, nil, "put error for test_invoke_module_failed", v, e)
	}
	v, e = drv.Create(params)
	if nil == v {
		testResult(t, drv, nil, "record not found", nil, e)
	} else {
		testResult(t, drv, nil, "record not found", v, e)
	}

	b, e := drv.Delete(params)
	testResult(t, drv, true, "", b, e)
}

func TestInvokeModuleAndCallback(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestInvokeModuleAndCallback"
	drv.Start()

	td := &TestDriver{get: "get test cb ok test1whj23", put: "put test cb ok test1whj23",
		create: false, delete: false, create_error: errutils.InternalError("create test cb ok test1whj23"),
		delete_error: errutils.InternalError("delete test cb ok test1whj23")}
	drv.drvMgr.Register("test_dumy_TestInvokeModuleAndCallback", td)

	defer func() {
		drv.Stop()
		drv.drvMgr.Unregister("test_dumy_TestInvokeModuleAndCallback")
	}()

	params := map[string]string{"schema": "test_invoke_module_and_callback", "dumy": "test_dumy_TestInvokeModuleAndCallback"}
	v, e := drv.Get(params)
	testResult(t, drv, map[string]interface{}{"value": "get test cb ok test1whj23"}, "", v, e)
	v, e = drv.Put(params)
	testResult(t, drv, map[string]interface{}{"value": "put test cb ok test1whj23"}, "", v, e)
	v, e = drv.Create(params)
	testResult(t, drv, map[string]interface{}{"value": false}, "create test cb ok test1whj23", v, e)
	b, e := drv.Delete(params)
	testResult(t, drv, false, "delete test cb ok test1whj23", b, e)
}

func TestInvokeModuleAndCallbackFailed(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestInvokeModuleAndCallbackFailed"
	drv.Start()

	td := &TestDriver{get_error: errutils.InternalError("get test cb ok test1whj23"), put_error: errutils.InternalError("put test cb ok test1whj23"),
		create: false, delete: false}
	drv.drvMgr.Register("test_dumy_TestInvokeModuleAndCallback", td)

	defer func() {
		drv.Stop()
		drv.drvMgr.Unregister("test_dumy_TestInvokeModuleAndCallback")
	}()

	params := map[string]string{"schema": "test_invoke_module_and_callback", "dumy": "test_dumy_TestInvokeModuleAndCallback"}
	v, e := drv.Get(params)
	if nil == v {
		testResult(t, drv, nil, "get test cb ok test1whj23", nil, e)
	} else {
		testResult(t, drv, nil, "get test cb ok test1whj23", v, e)
	}

	v, e = drv.Put(params)
	if nil == v {
		testResult(t, drv, nil, "put test cb ok test1whj23", nil, e)
	} else {
		testResult(t, drv, nil, "put test cb ok test1whj23", v, e)
	}
	v, e = drv.Create(params)
	testResult(t, drv, map[string]interface{}{"value": false}, "", v, e)
	b, e := drv.Delete(params)
	testResult(t, drv, false, "", b, e)
}

func TestDeliveryComplexBetweenGOAndLua(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestInvokeModuleAndCallbackFailed"
	drv.Start()

	td := &TestDriver{get: map[string]interface{}{"i8": int8(8), "i16": int16(16), "int32": int32(32), "int64": int64(64), "int": 128,
		"ui8": uint8(8), "ui16": uint16(16), "uint32": uint32(32), "uint64": uint64(64), "uint": uint(128),
		"string":       "string_value",
		"int_array":    []int{1, 2, 3, 4, 5},
		"string_array": []string{"1", "2", "3", "4", "5"},
		"object": map[string]interface{}{"i8": int8(8), "i16": int16(16), "int32": int32(32), "int64": int64(64), "int": 128,
			"ui8": uint8(8), "ui16": uint16(16), "uint32": uint32(32), "uint64": uint64(64), "uint": uint(128),
			"string":       "string_value",
			"int_array":    []int{1, 2, 3, 4, 5},
			"string_array": []string{"1", "2", "3", "4", "5"}}}}

	drv.drvMgr.Register("test_dumy_TestInvokeModuleAndCallback", td)

	defer func() {
		drv.Stop()
		drv.drvMgr.Unregister("test_dumy_TestInvokeModuleAndCallback")
	}()

	params := map[string]string{"schema": "test_invoke_module_and_callback", "dumy": "test_dumy_TestInvokeModuleAndCallback"}
	v, e := drv.Get(params)
	if nil != e {
		t.Errorf("execute failed - %s", e.Error())
		return
	}
	bytes1, err := json.Marshal(map[string]interface{}{"value": td.get})
	if nil != err {
		t.Errorf("excepted to json failed - %s", err.Error())
		return
	}
	j1 := make(map[string]interface{})
	err = json.Unmarshal(bytes1, &j1)
	if nil != err {
		t.Errorf("excepted parse json failed - %s", err.Error())
		return
	}
	bytes2, err := json.Marshal(v)
	if nil != err {
		t.Errorf("actual to json failed - %s", err.Error())
		return
	}
	j2 := make(map[string]interface{})
	err = json.Unmarshal(bytes2, &j2)
	if nil != err {
		t.Errorf("actual parse json failed - %s", err.Error())
		return
	}

	if !reflect.DeepEqual(j1, j2) {
		t.Log(string(bytes1))
		t.Log(string(bytes2))
		t.Error("j1 != j2")
	}

	fmt.Println(string(bytes1))
	fmt.Println(string(bytes2))
}

func TestInitScriptWithErrorSyntex(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.init_path = "test/lua_error_script.lua"
	err := drv.Start()
	if nil == err {
		t.Errorf("test start lua failed, except return a error, actual return success")
	} else if !strings.Contains(err.Error(), ": 'end' expected near <eof>") {
		t.Errorf("test start lua failed, excepted value contains ': 'end' expected near <eof>', actual value is " + err.Error())
	}
}

func doFunc(b bool, t *testing.T) {
	if b {
		defer func() {
			t.Error("it is failed")
		}()
	}
}

func TestDefer(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	doFunc(false, t)
}

func TestInitFiles(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggers(nil, func(s string) error { t.Log(s); return nil }, "", 0)
	drv.Name = "TestInitFiles"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_ok_init"}
	v, e := drv.Get(params)
	if nil != e {
		t.Error(e)
		t.FailNow()
	}

	s, ok := v["value"].(string)
	if !ok {
		t.Log(v, e)
		t.Errorf("return is not a string, %T", v)
		t.FailNow()
	}

	if "test init ok" != s {
		t.Log(v, e)
		t.Error("return != 'test init ok', it is %s", s)
		t.FailNow()
	}
}
