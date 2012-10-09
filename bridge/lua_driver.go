package main

// #cgo CFLAGS: -I ./include
// #cgo windows LDFLAGS: -llua52
// #cgo LDFLAGS: -lm -L ./lib
// #include <stdlib.h>
// #include "lua.h"
// #include "lualib.h"
// #include "lauxlib.h"
import "C"
import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"snmp"
	"sync"
	"time"
	"unsafe"
)

const (
	lua_init_script string = ``

	LUA_YIELD     int = C.LUA_YIELD
	LUA_ERRRUN        = C.LUA_ERRRUN
	LUA_ERRSYNTAX     = C.LUA_ERRSYNTAX
	LUA_ERRMEM        = C.LUA_ERRMEM
	LUA_ERRERR        = C.LUA_ERRERR
	LUA_ERRFILE       = C.LUA_ERRFILE
)

type LUA_CODE int

const (
	LUA_EXECUTE_END      LUA_CODE = 0
	LUA_EXECUTE_CONTINUE LUA_CODE = 1
	LUA_EXECUTE_FAILED   LUA_CODE = 2
)

type LuaDriver struct {
	snmp.Svc
	ls          *C.lua_State
	loop        *C.lua_State
	init_script string
	waitG       sync.WaitGroup
}

type Continuous struct {
	ls     *C.lua_State
	status LUA_CODE
	drv    string
	action string

	params map[string]string
	any    interface{}
	err    error
}

// func (svc *Svc) Set(onStart, onStop, onTimeout func()) {
// 	svc.onStart = onStart
// 	svc.onStop = onStop
// 	svc.onTimeout = onTimeout
// }
func NewLuaDriver(init_script string) *LuaDriver {
	driver := &LuaDriver{init_script: init_script}
	driver.Set(func() { driver.atStart() }, func() { driver.atStop() }, nil)
	return driver
}

func lua_init(ls *C.lua_State, default_script string) C.int {
	var cs *C.char
	defer func() {
		if nil != cs {
			C.free(unsafe.Pointer(cs))
		}
	}()
	wd, err := os.Getwd()
	if nil == err {
		pa := path.Join(wd, "lua_init.lua")
		if fileExists(pa) {
			cs = C.CString(pa)
			return C.luaL_loadfilex(ls, cs, nil)
		}
		log.Printf("LuaDriver: '%s' is not exist.", pa)
	}

	cs = C.CString(default_script)
	return C.luaL_loadstring(ls, cs)
}

func (driver *LuaDriver) atStart() {
	ls := C.luaL_newstate()
	defer func() {
		if nil != ls {
			C.lua_close(ls)
		}
	}()
	C.luaL_openlibs(ls)

	if "" == driver.init_script {
		driver.init_script = lua_init_script
	}
	ret := lua_init(ls, driver.init_script)

	if LUA_ERRFILE == ret {
		panic("'lua_init.lua' read fail")
	} else if 0 != ret {
		panic(getError(ls, ret, "load 'lua_init.lua' failed").Error())
	}

	ret = C.lua_resume(ls, nil, 0)
	if 0 != int(ret) {
		panic(getError(ls, ret, "launch main fiber failed").Error())
	}
	loop := C.lua_tothread(ls, -1)
	if nil == loop {
		panic(getError(ls, ret, "launch loop fiber failed").Error())
	}

	driver.loop = loop
	driver.ls = ls
	ls = nil
}

func (driver *LuaDriver) atStop() {
	pushString(driver.loop, "__exit__")
	if nil == driver.loop || nil == driver.ls {
		panic("stop failed.")
	}
	ret := C.lua_resume(driver.loop, driver.ls, 1)
	if 0 != ret {
		if 0 == LUA_YIELD {
			panic("'lua_init.lua' is PAUSE.")
		} else {
			panic(getError(driver.loop, ret, "stop main fiber failed").Error())
		}
	}

	driver.waitG.Wait()
	C.lua_close(driver.loop)
	C.lua_close(driver.ls)

	driver.loop = nil
	driver.ls = nil
}

func (driver *LuaDriver) executeTask(drv, action string, params map[string]string) (ret interface{}, err error) {
	err = errors.New("not implemented")
	return
}

func (driver *LuaDriver) newContinuous(action string, params map[string]string) *Continuous {
	pushString(driver.loop, action)
	pushTable(driver.loop, params)

	ret := C.lua_resume(driver.loop, driver.ls, 2)
	if LUA_YIELD != int(ret) {
		if 0 == ret {
			return &Continuous{status: LUA_EXECUTE_FAILED,
				err: errors.New("'lua_init.lua' is directly exited.")}
		} else {
			return &Continuous{status: LUA_EXECUTE_FAILED,
				err: getError(driver.loop, ret, "switch to main fiber failed")}
		}
	}

	new_th := C.lua_tothread(driver.loop, -1)
	if nil == new_th {
		return &Continuous{status: LUA_EXECUTE_FAILED,
			err: errors.New("main fiber return value by yeild is not lua_Status type")}
	}

	ct := &Continuous{status: LUA_EXECUTE_FAILED, ls: new_th}
	ct = driver.executeContinuous(ct)

	if LUA_EXECUTE_CONTINUE == ct.status {
		driver.waitG.Add(1)
	}
	return ct
}

func (driver *LuaDriver) executeContinuous(ct *Continuous) *Continuous {
	switch ret := C.lua_status(ct.ls); ret {
	case C.LUA_YIELD:
		ct.status = LUA_EXECUTE_CONTINUE
		ct.drv = toString(ct.ls, -3)
		ct.action = toString(ct.ls, -2)
		ct.params = toParams(ct.ls, -1)
	case 0:
		ct.status = LUA_EXECUTE_END
		ct.any = toAny(ct.ls, -2)
		ct.err = toError(ct.ls, -1)
		C.lua_close(ct.ls)
	default:
		ct.status = LUA_EXECUTE_FAILED
		ct.err = getError(ct.ls, ret, "script execute failed - ")
		C.lua_close(ct.ls)
	}
	return ct
}

func toContinuous(values []interface{}) (ct *Continuous, err error) {

	if 2 <= len(values) && nil != values[1] {
		err = values[1].(error)
	}

	if nil != values[0] {
		var ok bool = false
		ct, ok = values[0].(*Continuous)
		if !ok {
			err = snmp.NewTwinceError(err, fmt.Errorf("oooooooo! It is not a Continuous - %v", values[0]))
		}
	} else if nil == err {
		err = errors.New("oooooooo! return a nil")
	}
	return
}

func (driver *LuaDriver) invoke(action string, params map[string]string) (interface{}, error) {
	t := 5 * time.Minute
	old := time.Now()

	values := driver.SafelyCall(t, func() *Continuous {
		return driver.newContinuous(action, params)
	})
	ct, err := toContinuous(values)
	if nil != err {
		if nil != ct && LUA_EXECUTE_CONTINUE == ct.status {
			driver.waitG.Done()
		}
		return nil, err
	}

	if LUA_EXECUTE_CONTINUE == ct.status {
		defer func() {
			driver.waitG.Done()
		}()

		for {
			seconds := (time.Now().Second() - old.Second())
			t -= (time.Duration(seconds) * time.Second)
			values := driver.SafelyCall(t, func() *Continuous {
				return driver.executeContinuous(ct)
			})

			ct, err = toContinuous(values)
			if nil != err {
				return nil, err
			}
			if LUA_EXECUTE_CONTINUE != ct.status {
				break
			}
		}
	}

	if LUA_EXECUTE_END == ct.status {
		return ct.any, ct.err
	}
	return nil, ct.err
}

func (driver *LuaDriver) invokeAndReturnBool(action string, params map[string]string) (bool, error) {
	ret, err := driver.invoke(action, params)
	if nil == ret {
		return false, err
	}
	b, ok := ret.(bool)
	if !ok {
		panic(fmt.Sprintf("type of result is not bool type - %v", b))
	}

	return b, err
}
func (driver *LuaDriver) Get(params map[string]string) (interface{}, error) {
	return driver.invoke("get", params)
}

func (driver *LuaDriver) Put(params map[string]string) (interface{}, error) {
	return driver.invoke("put", params)
}

func (driver *LuaDriver) Create(params map[string]string) (bool, error) {
	return driver.invokeAndReturnBool("create", params)
}

func (driver *LuaDriver) Delete(params map[string]string) (bool, error) {
	return driver.invokeAndReturnBool("delete", params)
}
