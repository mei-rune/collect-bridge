package lua_binding

import (
	"commons"
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

func checkResult(t *testing.T, excepted, old, actual interface{}, err error, msg string) {
	if nil != err {
		t.Errorf(msg+err.Error()+" - %v", old)
	} else if excepted != actual {
		t.Errorf(msg+" %v !=  %v - %v", excepted, actual, old)
	}
}

func checkArray(t *testing.T, old interface{}) {
	var res interface{}
	var err error

	array, err := commons.AsArray(old)
	if nil != err {
		t.Errorf("test array - "+err.Error()+" - %v", old)
	} else {
		res, err = commons.AsInt(array[0])
		checkResult(t, int(1), array[0], res, err, "test int in array - ")

		res, err = commons.AsUint(array[1])
		checkResult(t, uint(2), array[1], res, err, "test uint in array - ")

		res, err = commons.AsString(array[2])
		checkResult(t, "s1", array[2], res, err, "test string in array - ")
	}
}

func checkMap(t *testing.T, old interface{}) {

	var res interface{}
	var err error

	assoc, err := commons.AsMap(old)
	if nil != err {
		t.Errorf("test map - "+err.Error()+" - %v", old)
	} else {
		fmt.Print(assoc)

		res, err = commons.AsInt(assoc["a1"])
		checkResult(t, int(1), assoc["a1"], res, err, "test int in map - ")

		res, err = commons.AsUint(assoc["a2"])
		checkResult(t, uint(2), assoc["a2"], res, err, "test uint in map - ")

		res, err = commons.AsString(assoc["a3"])
		checkResult(t, "s3", assoc["a3"], res, err, "test string in array - ")
	}
}

func TestParams(t *testing.T) {

	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
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
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
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
	checkResult(t, nil, old, old, nil, "test nil - ")

	old = pushAnyTest(drv, true)
	if nil == old {
		panic("nil == old")
	}
	res, err = commons.AsBool(old)
	checkResult(t, true, old, res, err, "test bool - ")

	old = pushAnyTest(drv, int8(8))
	res, err = commons.AsInt8(old)
	checkResult(t, int8(8), old, res, err, "test int8 - ")

	old = pushAnyTest(drv, int16(16))
	res, err = commons.AsInt16(old)
	checkResult(t, int16(16), old, res, err, "test int16 - ")

	old = pushAnyTest(drv, int32(32))
	res, err = commons.AsInt32(old)
	checkResult(t, int32(32), old, res, err, "test int32 - ")

	old = pushAnyTest(drv, int64(64))
	res, err = commons.AsInt64(old)
	checkResult(t, int64(64), old, res, err, "test int64 - ")

	old = pushAnyTest(drv, uint8(98))
	res, err = commons.AsUint8(old)
	checkResult(t, uint8(98), old, res, err, "test uint8 - ")

	old = pushAnyTest(drv, uint16(916))
	res, err = commons.AsUint16(old)
	checkResult(t, uint16(916), old, res, err, "test uint16 - ")

	old = pushAnyTest(drv, uint32(932))
	res, err = commons.AsUint32(old)
	checkResult(t, uint32(932), old, res, err, "test uint32 - ")

	old = pushAnyTest(drv, uint64(964))
	res, err = commons.AsUint64(old)
	checkResult(t, uint64(964), old, res, err, "test uint64 - ")

	old = pushAnyTest(drv, uint(123))
	res, err = commons.AsUint(old)
	checkResult(t, uint(123), old, res, err, "test uint - ")

	old = pushAnyTest(drv, int(223))
	res, err = commons.AsInt(old)
	checkResult(t, int(223), old, res, err, "test int - ")

	old = pushAnyTest(drv, "aa")
	res, err = commons.AsString(old)
	checkResult(t, "aa", old, res, err, "test string - ")

	old = pushAnyTest(drv, []interface{}{int8(1), uint8(2), "s1"})
	checkArray(t, old)

	old = pushAnyTest(drv, map[string]interface{}{"a1": 1, "a2": uint(2), "a3": "s3"})
	checkMap(t, old)

	old = pushAnyTest(drv, []interface{}{int8(1), uint8(2), "s1", map[string]interface{}{"a1": 1, "a2": uint(2), "a3": "s3"}})
	checkArray(t, old)

	array, err := commons.AsArray(old)
	if nil == err {
		t.Log("test map in array")
		checkMap(t, array[3])
		t.Log("test map in array is ok")
	}

	old = pushAnyTest(drv, map[string]interface{}{"a1": 1, "a2": uint(2), "a3": "s3", "a4": []interface{}{int8(1), uint8(2), "s1"}})
	checkMap(t, old)

	assoc, err := commons.AsMap(old)
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
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "test_invoke"
	drv.init_path = "test/lua_init_test.lua"
	drv.Start()

	if !drv.IsRunning() {
		t.Errorf("TestInvoke - start failed")
		return
	}

	v := drv.Get(nil)
	if v.HasError() {
		t.Errorf("execute get failed, " + v.ErrorMessage())
	} else if s, _ := v.Value().AsString(); "test ok" != s {
		t.Errorf("execute get failed, excepted value is 'ok', actual is %v", v)
	} else {
		t.Log("execute get ok")
	}
	drv.Stop()
}

func TestInvokeScript(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestInvokeScript"
	drv.Start()

	if !drv.IsRunning() {
		t.Errorf("TestInvokeScript - start failed")
		return
	}

	params := map[string]string{"schema": "script", "script": "return {value= action..' ok'}"}
	v := drv.Get(params)
	if v.HasError() {
		t.Errorf("execute get failed, " + v.ErrorMessage())
	} else if s, _ := v.Value().AsString(); "get ok" != s {
		t.Errorf("execute get failed, excepted value is 'ok', actual is %v", v)
	} else {
		t.Log("execute get ok")
	}
	drv.Stop()
}

func TestInvokeScriptFailed(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestInvokeScriptFailed"
	drv.Start()

	if !drv.IsRunning() {
		t.Errorf("TestInvokeScriptFailed - start failed")
		return
	}

	params := map[string]string{"schema": "script", "script": "aa"}
	res := drv.Get(params)
	if !res.HasError() {
		t.Errorf("execute get failed, except return error, actual return ok")
	} else if !strings.Contains(res.ErrorMessage(), "syntax error near <eof>") {
		t.Errorf("execute get failed, except error contains 'syntax error near <eof>', actual return - " + res.ErrorMessage())
	}
	drv.Stop()
	drv.drvMgr.Unregister("test")
}

func TestInvokeAndCallback(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "TestInvokeAndCallback", 0)
	drv.Name = "TestInvokeAndCallback"
	drv.Start()

	if !drv.IsRunning() {
		t.Errorf("TestInvokeAndCallback - start failed")
		return
	}

	td := &commons.DefaultDrv{GetValue: "get12", PutValue: "put12", CreateValue: true, DeleteValue: true}
	drv.drvMgr.Register("test_dumy", td)

	defer func() {
		drv.Stop()
		drv.drvMgr.Unregister("test_dumy")
	}()

	params := map[string]string{"schema": "script", "script": "mj.log(mj.DEBUG, 'log a test log.')\nreturn mj.execute('test_dumy', action, params)"}
	res := drv.Get(params)
	if res.HasError() {
		t.Errorf("execute get failed, " + res.ErrorMessage())
	} else if s, _ := res.Value().AsString(); "get12" != s {
		t.Errorf("execute get failed, excepted value is 'get12', actual is %v", res.Value())
	} else {
		t.Log("execute get ok")
	}
}

func testResult(t *testing.T, drv *LuaDriver, excepted_value interface{}, excepted_error string, actual_result commons.Result) {
	if !actual_result.HasError() {
		if "" != excepted_error {
			t.Errorf("execute failed, excepted error is %v, actual error is nil", excepted_error)
		}
	} else if actual_result.ErrorMessage() != excepted_error {
		t.Errorf("execute failed, excepted error is %v, actual error is %v", excepted_error, actual_result.ErrorMessage())
	}

	if nil == actual_result.Value().AsInterface() {
		if nil != excepted_value {
			t.Errorf("execute failed, excepted value is %v, actual value is nil", excepted_value)
		}
	} else if !reflect.DeepEqual(actual_result.Value().AsInterface(), excepted_value) {
		t.Errorf("execute failed, excepted value is '%v', actual value is %v", excepted_value, actual_result.Value().AsInterface())
	} else {
		t.Log("execute ok")
	}
}

func TestInvokeModule(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestInvokeModule"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_invoke_module"}
	v := drv.Get(params)
	testResult(t, drv, "get test ok test1whj23", "", v)
	v = drv.Put(params)
	testResult(t, drv, "put test ok test1whj23", "", v)
	v = drv.Create(params)
	testResult(t, drv, "2328", "create test ok test1whj23", v)
	v = drv.Delete(params)
	testResult(t, drv, false, "delete test ok test1whj23", v)
}

func TestInvokeModuleFailed(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, nil)
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestInvokeModuleFailed"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_invoke_module_failed"}
	v := drv.Get(params)
	testResult(t, drv, nil, "get error for test_invoke_module_failed", v)

	v = drv.Put(params)
	testResult(t, drv, nil, "put error for test_invoke_module_failed", v)

	v = drv.Create(params)
	testResult(t, drv, nil, "record not found", v)

	v = drv.Delete(params)
	testResult(t, drv, nil, "delete failed", v)
}

func TestInvokeModuleAndCallback(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestInvokeModuleAndCallback"
	drv.Start()

	td := &commons.DefaultDrv{GetValue: "get test cb ok test1whj23", PutValue: "put test cb ok test1whj23",
		CreateValue: false, DeleteValue: false, CreateCode: commons.InternalErrorCode,
		CreateErr:  "create test cb ok test1whj23",
		DeleteCode: commons.InternalErrorCode,
		DeleteErr:  "delete test cb ok test1whj23"}
	drv.drvMgr.Register("test_dumy_TestInvokeModuleAndCallback", td)

	defer func() {
		drv.Stop()
		drv.drvMgr.Unregister("test_dumy_TestInvokeModuleAndCallback")
	}()

	params := map[string]string{"schema": "test_invoke_module_and_callback", "dumy": "test_dumy_TestInvokeModuleAndCallback"}
	v := drv.Get(params)
	testResult(t, drv, "get test cb ok test1whj23", "", v)
	v = drv.Put(params)
	testResult(t, drv, "put test cb ok test1whj23", "", v)
	v = drv.Create(params)
	testResult(t, drv, false, "create test cb ok test1whj23", v)
	b := drv.Delete(params)
	testResult(t, drv, false, "delete test cb ok test1whj23", b)
}

func TestInvokeModuleAndCallbackFailed(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestInvokeModuleAndCallbackFailed"
	drv.Start()

	td := &commons.DefaultDrv{GetCode: commons.InternalErrorCode, GetErr: "get test cb ok test1whj23",
		PutCode: commons.InternalErrorCode, PutErr: "put test cb ok test1whj23",
		CreateValue: false, DeleteValue: false}
	drv.drvMgr.Register("test_dumy_TestInvokeModuleAndCallback", td)

	defer func() {
		drv.Stop()
		drv.drvMgr.Unregister("test_dumy_TestInvokeModuleAndCallback")
	}()

	params := map[string]string{"schema": "test_invoke_module_and_callback", "dumy": "test_dumy_TestInvokeModuleAndCallback"}
	v := drv.Get(params)
	testResult(t, drv, nil, "get test cb ok test1whj23", v)

	v = drv.Put(params)
	testResult(t, drv, nil, "put test cb ok test1whj23", v)

	v = drv.Create(params)
	testResult(t, drv, false, "", v)

	b := drv.Delete(params)
	testResult(t, drv, false, "", b)
}

func TestDeliveryComplexBetweenGOAndLua(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestInvokeModuleAndCallbackFailed"
	drv.Start()

	td := &commons.DefaultDrv{GetValue: map[string]interface{}{"i8": int8(8), "i16": int16(16), "int32": int32(32), "int64": int64(64), "int": 128,
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
	v := drv.Get(params)
	if v.HasError() {
		t.Errorf("execute failed - %s", v.ErrorMessage())
		return
	}
	bytes1, err := json.Marshal(td.GetValue)
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
	bytes2, err := json.Marshal(v.InterfaceValue())
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
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
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
	drv.InitLoggerWithCallback(func(s []byte) { t.Log(string(s)) }, "", 0)
	drv.Name = "TestInitFiles"
	drv.Start()
	defer func() {
		drv.Stop()
	}()
	params := map[string]string{"schema": "test_ok_init"}
	v := drv.Get(params)
	if v.HasError() {
		t.Error(v.ErrorMessage())
		t.FailNow()
	}

	s, ok := v.Value().AsString()
	if nil != ok {
		t.Errorf("return is not a string, %T", v)
		t.FailNow()
	}

	if "test init ok" != s {
		t.Error("return != 'test init ok', it is %s", s)
		t.FailNow()
	}
}
