package lua_binding

import (
	"commons"
	"commons/errutils"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"
)

func BenchmarkInvokeArgumentsByStack(b *testing.B) {

	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggerWithCallback(func(s []byte) { b.Log(string(s)) }, "", 0)
	drv.Name = "BenchmarkInvokeArgumentsByStack"

	drv.init_path = "test/lua_init_test_push_bench.lua"
	err := drv.Start()
	if nil != err {
		b.Error(err)
		return
	}

	defer func() {
		drv.Stop()
	}()

	params := map[string]string{"schema": "test_invoke_module", "dumy": "test_dumy_TestInvokeModuleAndCallback"}

	now := time.Now()
	b.ResetTimer()
	for i := 0; i < 10000; i++ {
		PushString(drv, "test_invoke_module2")
		PushParams(drv, params)
		ResumeLuaFiber(drv, 2)
		_, err := ToAny(drv, -1)
		if nil != err {
			b.Error(err)
			return
		}
	}

	b.StopTimer()
	fmt.Printf("BenchmarkInvokeArgumentsByStack time is %v\n", time.Now().Sub(now))
}
func BenchmarkInvokeArgumentsByJSON(b *testing.B) {

	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggerWithCallback(func(s []byte) { b.Log(string(s)) }, "", 0)
	drv.Name = "BenchmarkInvokeArgumentsByJSON"

	drv.init_path = "test/lua_init_test_json_bench.lua"
	err := drv.Start()
	if nil != err {
		b.Error(err)
		return
	}

	defer func() {
		drv.Stop()
	}()
	result := map[string]string{}
	params := map[string]string{"schema": "test_invoke_module", "dumy": "test_dumy_TestInvokeModuleAndCallback"}

	now := time.Now()
	b.ResetTimer()
	for i := 0; i < 10000; i++ {
		j1, err := json.Marshal(params)
		if nil != err {
			b.Error(err)
			return
		}
		PushString(drv, "test_invoke_module2")
		PushString(drv, string(j1))
		ResumeLuaFiber(drv, 2)
		j2, err := ToString(drv, -1)
		if nil != err {
			b.Error(err)
			return
		}

		err = json.Unmarshal([]byte(j2), &result)
		if nil != err {
			b.Error(err)
			return
		}
	}

	b.StopTimer()
	fmt.Printf("BenchmarkInvokeArgumentsByJSON time is %v\n", time.Now().Sub(now))
}

func BenchmarkGet(b *testing.B) {

	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggerWithCallback(func(s []byte) { b.Log(string(s)) }, "", 0)
	drv.Name = "BenchmarkGet"
	drv.Start()

	td := &TestDriver{get: "get test cb ok test1whj23", put: "put test cb ok test1whj23",
		create: false, delete: false, create_error: errutils.InternalError("create test cb ok test1whj23"),
		delete_error: errutils.InternalError("delete test cb ok test1whj23")}
	drv.drvMgr.Register("test_dumy_TestInvokeModuleAndCallback", td)

	defer func() {
		drv.Stop()
		drv.drvMgr.Unregister("test_dumy_TestInvokeModuleAndCallback")
	}()

	params := map[string]string{"schema": "test_invoke_module", "dumy": "test_dumy_TestInvokeModuleAndCallback"}

	now := time.Now()
	b.ResetTimer()
	for i := 0; i < 10000; i++ {
		drv.Get(params)
	}
	b.StopTimer()
	fmt.Printf("BenchmarkGet time is %v\n", time.Now().Sub(now))
}

func BenchmarkGetWithCallback(b *testing.B) {

	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggerWithCallback(func(s []byte) { b.Log(string(s)) }, "", 0)
	drv.Name = "BenchmarkGetWithCallback"
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

	now := time.Now()
	b.ResetTimer()
	for i := 0; i < 10000; i++ {
		drv.Get(params)
	}
	b.StopTimer()
	fmt.Printf("BenchmarkGetWithCallback time is %v\n", time.Now().Sub(now))
}

func BenchmarkDirectGet(b *testing.B) {

	log.SetFlags(log.Flags() | log.Lshortfile)

	drv := NewLuaDriver(1*time.Second, commons.NewDriverManager())
	drv.InitLoggerWithCallback(func(s []byte) { b.Log(string(s)) }, "", 0)
	drv.Name = "BenchmarkDirectGet"
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

	now := time.Now()
	b.ResetTimer()
	for i := 0; i < 10000; i++ {
		d, _ := drv.drvMgr.Connect("test_dumy_TestInvokeModuleAndCallback")
		d.Get(params)
	}
	b.StopTimer()
	fmt.Printf("BenchmarkDirectGet time is %v\n", time.Now().Sub(now))
}
