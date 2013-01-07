package lua_binding

// #cgo windows CFLAGS: -DLUA_COMPAT_ALL -DLUA_COMPAT_ALL -I ./include
// #cgo windows LDFLAGS: -L ./lib -llua52 -lm
// #cgo linux CFLAGS: -DLUA_USE_LINUX -DLUA_COMPAT_ALL
// #cgo linux LDFLAGS: -L. -llua52 -ldl  -lm
// #include <stdlib.h>
// #include "lua.h"
// #include "lualib.h"
// #include "lauxlib.h"
import "C"
import (
	"commons"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

const (
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
	LUA_EXECUTE_YIELD    LUA_CODE = 2
	LUA_EXECUTE_FAILED   LUA_CODE = 3
)

type NativeMethod struct {
	Name     string
	Read     func(drv *LuaDriver, ctx *Continuous)
	Write    func(drv *LuaDriver, ctx *Continuous) (int, error)
	Callback func(drv *LuaDriver, ctx *Continuous)
}

var (
	method_init_lua = &NativeMethod{
		Name:     "method_init_lua",
		Read:     nil,
		Write:    nil,
		Callback: nil}
	method_exit_lua = &NativeMethod{
		Name: "method_exit_lua",
		Read: nil,
		Write: func(drv *LuaDriver, ctx *Continuous) (int, error) {
			err := ctx.PushStringParam("__exit__")
			if nil != err {
				return -1, err
			}
			return 1, err
		},
		Callback: nil}

	method_missing = &NativeMethod{
		Name:     "method_missing",
		Read:     nil,
		Write:    writeCallResult,
		Callback: nil}
)

type LuaDriver struct {
	commons.Svc
	init_path string
	LS        *C.lua_State
	waitG     sync.WaitGroup

	methods        map[string]*NativeMethod
	method_missing *NativeMethod
}

type Continuous struct {
	LS     *C.lua_State
	status LUA_CODE
	method *NativeMethod

	on_end func(drv *LuaDriver, ctx *Continuous)

	Error       error
	IntValue    int
	StringValue string
	Params      map[string]string
	Any         interface{}
}

func (self *Continuous) clear() {
	self.method = nil
	self.Error = nil
	self.IntValue = 0
	self.StringValue = ""
	self.Params = nil
	self.Any = nil
}

func (self *Continuous) ToErrorParam(idx int) error {
	return toError(self.LS, C.int(idx))
}

func (self *Continuous) ToAnyParam(idx int) (interface{}, error) {
	return toAny(self.LS, C.int(idx))
}

func (self *Continuous) ToParamsParam(idx int) (map[string]string, error) {
	return toParams(self.LS, C.int(idx))
}

func (self *Continuous) ToStringParam(idx int) (string, error) {
	return toString(self.LS, C.int(idx))
}

func (self *Continuous) ToIntParam(idx int) (int, error) {
	return toInteger(self.LS, C.int(idx))
}

func (self *Continuous) PushAnyParam(any interface{}) error {
	pushAny(self.LS, any)
	return nil
}

func (self *Continuous) PushParamsParam(params map[string]string) error {
	pushParams(self.LS, params)
	return nil
}

func (self *Continuous) PushStringParam(s string) error {
	pushString(self.LS, s)
	return nil
}

func (self *Continuous) PushErrorParam(e error) error {
	pushError(self.LS, e)
	return nil
}

func readCallArguments(drv *LuaDriver, ctx *Continuous) {
	ctx.StringValue, ctx.Error = ctx.ToStringParam(2)
	if nil != ctx.Error {
		return
	}
	ctx.Params, ctx.Error = ctx.ToParamsParam(3)
}

func writeCallResult(drv *LuaDriver, ctx *Continuous) (int, error) {
	err := ctx.PushAnyParam(ctx.Any)
	if nil != err {
		return -1, err
	}
	err = ctx.PushErrorParam(ctx.Error)
	if nil != err {
		return -1, err
	}
	return 2, nil
}

func readActionResult(drv *LuaDriver, ctx *Continuous) {
	ctx.Any, ctx.Error = ctx.ToAnyParam(-2)
	if nil != ctx.Error {
		return
	}
	ctx.Error = ctx.ToErrorParam(-1)
}

func writeActionArguments(drv *LuaDriver, ctx *Continuous) (int, error) {

	err := ctx.PushStringParam(ctx.StringValue)
	if nil != err {
		return -1, err
	}
	err = ctx.PushParamsParam(ctx.Params)
	if nil != err {
		return -1, err
	}
	return 2, nil
}

// func (svc *Svc) Set(onStart, onStop, onTimeout func()) {
//	svc.onStart = onStart
//	svc.onStop = onStop
//	svc.onTimeout = onTimeout
// }
func NewLuaDriver() *LuaDriver {
	driver := &LuaDriver{}
	driver.Name = "lua_driver"
	driver.methods = make(map[string]*NativeMethod)
	driver.Set(func() { driver.atStart() }, func() { driver.atStop() }, nil)
	err := driver.CallbackWith(&NativeMethod{
		Name:  "get",
		Read:  readCallArguments,
		Write: writeCallResult,
		Callback: func(lua *LuaDriver, ctx *Continuous) {
			drv, ok := commons.Connect(ctx.StringValue)
			if !ok {
				ctx.Error = fmt.Errorf("driver '%s' is not exists.", ctx.StringValue)
				return
			}

			ctx.Any, ctx.Error = drv.Get(ctx.Params)
		}}, &NativeMethod{
		Name:  "put",
		Read:  readCallArguments,
		Write: writeCallResult,
		Callback: func(lua *LuaDriver, ctx *Continuous) {
			drv, ok := commons.Connect(ctx.StringValue)
			if !ok {
				ctx.Error = fmt.Errorf("driver '%s' is not exists.", ctx.StringValue)
				return
			}

			ctx.Any, ctx.Error = drv.Put(ctx.Params)
		}}, &NativeMethod{
		Name:  "create",
		Read:  readCallArguments,
		Write: writeCallResult,
		Callback: func(lua *LuaDriver, ctx *Continuous) {
			drv, ok := commons.Connect(ctx.StringValue)
			if !ok {
				ctx.Error = fmt.Errorf("driver '%s' is not exists.", ctx.StringValue)
				return
			}

			ctx.Any, ctx.Error = drv.Create(ctx.Params)
		}}, &NativeMethod{
		Name:  "delete",
		Read:  readCallArguments,
		Write: writeCallResult,
		Callback: func(lua *LuaDriver, ctx *Continuous) {
			drv, ok := commons.Connect(ctx.StringValue)
			if !ok {
				ctx.Error = fmt.Errorf("driver '%s' is not exists.", ctx.StringValue)
				return
			}

			ctx.Any, ctx.Error = drv.Delete(ctx.Params)
		}}, &NativeMethod{
		Name: "log",
		Read: func(drv *LuaDriver, ctx *Continuous) {
			ctx.IntValue, _ = ctx.ToIntParam(2)
			ctx.StringValue, _ = ctx.ToStringParam(3)
		},
		Write: func(drv *LuaDriver, ctx *Continuous) (int, error) {
			switch {
			case 9000 >= ctx.IntValue:
				drv.DEBUG.Print(ctx.StringValue)
			case 6000 >= ctx.IntValue:
				drv.INFO.Print(ctx.StringValue)
			case 4000 >= ctx.IntValue:
				drv.WARN.Print(ctx.StringValue)
			case 2000 >= ctx.IntValue:
				drv.ERROR.Print(ctx.StringValue)
			case 1000 >= ctx.IntValue:
				drv.FATAL.Print(ctx.StringValue)
			case 0 >= ctx.IntValue:
				drv.INFO.Print(ctx.StringValue)
			}
			return 0, nil
		},
		Callback: nil}, &NativeMethod{
		Name: "io_ext.enumerate_files",
		Read: func(drv *LuaDriver, ctx *Continuous) {
			ctx.StringValue, ctx.Error = ctx.ToStringParam(2)
		},
		Write: writeCallResult,
		Callback: func(lua *LuaDriver, ctx *Continuous) {
			if nil != ctx.Error {
				ctx.Any = nil
				return
			}
			ctx.Any, ctx.Error = commons.EnumerateFiles(ctx.StringValue)
		}})

	if nil != err {
		log.Panicln(err)
	}
	return driver
}

func (self *LuaDriver) CallbackWith(methods ...*NativeMethod) error {
	for _, m := range methods {
		if nil == m {
			return nil
		}
		if "" == m.Name {
			return errors.New("'name' is empty.")
		}
		if nil == m.Callback && nil == m.Write {
			return errors.New("'callback' of '" + m.Name + "' is nil.")
		}
		if _, ok := self.methods[m.Name]; ok {
			return errors.New("'" + m.Name + "' is already exists.")
		}
		if nil != self.DEBUG {
			self.DEBUG.Printf("register function '%s'", m.Name)
		} else {
			log.Printf("register function '%s'\n", m.Name)
		}
		self.methods[m.Name] = m
	}
	return nil
}

func (driver *LuaDriver) lua_init(ls *C.lua_State) C.int {
	var cs *C.char
	defer func() {
		if nil != cs {
			C.free(unsafe.Pointer(cs))
		}
	}()

	pushString(ls, runtime.GOARCH)

	cs = C.CString("__mj_arch")
	C.lua_setglobal(ls, cs)
	C.free(unsafe.Pointer(cs))
	cs = nil

	cs = C.CString("__mj_os")
	pushString(ls, runtime.GOOS)
	C.lua_setglobal(ls, cs)
	C.free(unsafe.Pointer(cs))
	cs = nil

	if len(os.Args) > 0 {
		pa := path.Base(os.Args[0])
		if fileExists(pa) {
			pushString(ls, pa)
			cs = C.CString("__mj_execute_directory")
			C.lua_setglobal(ls, cs)
			C.free(unsafe.Pointer(cs))
			cs = nil
		}
	}
	wd, err := os.Getwd()
	if nil == err {
		pushString(ls, wd)
		cs = C.CString("__mj_work_directory")
		C.lua_setglobal(ls, cs)
		C.free(unsafe.Pointer(cs))
		cs = nil

		pa := path.Join(wd, driver.init_path)
		if fileExists(pa) {
			cs = C.CString(pa)
			return C.luaL_loadfilex(ls, cs, nil)
		}
		driver.INFO.Printf("LuaDriver: '%s' is not exist.", pa)
	}

	if nil != cs {
		C.free(unsafe.Pointer(cs))
	}
	cs = C.CString(lua_init_script)
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

	if "" == driver.init_path {
		driver.init_path = "core.lua"
	}

	ret := driver.lua_init(ls)

	if LUA_ERRFILE == ret {
		driver.FATAL.Print("'" + driver.init_path + "' read fail")
	} else if 0 != ret {
		driver.FATAL.Print(getError(ls, ret, "load '"+driver.init_path+"' failed").Error())
	}

	ctx := &Continuous{LS: ls, method: method_init_lua}

	ctx = driver.eval(ctx)
	for LUA_EXECUTE_CONTINUE == ctx.status {
		if nil != ctx.method && nil != ctx.method.Callback {
			ctx.method.Callback(driver, ctx)
		}
		ctx = driver.eval(ctx)
	}

	if LUA_EXECUTE_YIELD != ctx.status {
		driver.FATAL.Print("launch main fiber failed, " + ctx.Error.Error())
	}

	driver.LS = ls
	ls = nil
	driver.INFO.Print("driver is started!")
}

func (driver *LuaDriver) atStop() {
	if nil == driver.LS {
		return
	}

	ret := C.lua_status(driver.LS)
	if C.LUA_YIELD != ret {
		driver.FATAL.Print(getError(driver.LS, ret, "stop main fiber failed, status is error").Error())
	}

	ctx := &Continuous{LS: driver.LS, method: method_exit_lua}

	ctx = driver.eval(ctx)
	for LUA_EXECUTE_CONTINUE == ctx.status {
		if nil != ctx.method && nil != ctx.method.Callback {
			ctx.method.Callback(driver, ctx)
		}
		ctx = driver.eval(ctx)
	}

	if LUA_EXECUTE_END != ctx.status {
		driver.FATAL.Print("stop main fiber failed," + ctx.Error.Error())
	}

	driver.INFO.Print("wait for all fibers to exit!")
	driver.waitG.Wait()
	driver.INFO.Print("all fibers is exited!")

	C.lua_close(driver.LS)
	driver.LS = nil
	driver.INFO.Print("driver is exited!")
}

func (self *LuaDriver) eval(ctx *Continuous) *Continuous {
	var ret C.int = 0
	var argc int = -1
	var ok bool = false
	var from *C.lua_State = nil

	if nil == ctx.LS {
		panic("aaaaa")
	}
	ls := ctx.LS
	if ls != self.LS {
		from = self.LS
	}

	for {
		if nil != ctx.method && nil != ctx.method.Write {
			argc, ctx.Error = ctx.method.Write(self, ctx)
			if nil != ctx.Error {
				ctx.status = LUA_EXECUTE_FAILED
				ctx.IntValue = C.LUA_ERRERR
				ctx.Error = errors.New("push arguments failed - " + ctx.Error.Error())
				return ctx
			}
		} else {
			argc = 0
		}

		ret = C.lua_resume(ls, from, C.int(argc))

		ctx.clear()

		switch ret {
		case 0:
			ctx.status = LUA_EXECUTE_END
			if nil != ctx.on_end {
				ctx.on_end(self, ctx)
			}
			// There is no explicit function to close or to destroy a thread. Threads are
			// subject to garbage collection, like any Lua object. 
			return ctx
		case C.LUA_YIELD:
			ctx.status = LUA_EXECUTE_YIELD
			if 0 == C.lua_gettop(ls) {
				ctx.IntValue = int(ret)
				ctx.Error = errors.New("script execute failed - return arguments is empty.")
				return ctx
			}

			action, err := toString(ls, 1)
			if nil != err {
				ctx.IntValue = int(ret)
				ctx.Error = errors.New("script execute failed, read action failed, " + err.Error())
				return ctx
			}
			ctx.method, ok = self.methods[action]
			if !ok {
				ctx.Error = fmt.Errorf("unsupport action '%s'", action)
				ctx.Any = nil
				ctx.method = method_missing
			}

			if nil != ctx.method.Read {
				ctx.method.Read(self, ctx)
			}

			if nil != ctx.method.Callback {
				ctx.status = LUA_EXECUTE_CONTINUE
				return ctx
			}
		default:
			ctx.status = LUA_EXECUTE_FAILED
			ctx.IntValue = int(ret)
			ctx.Error = getError(ls, ret, "script execute failed")
			// There is no explicit function to close or to destroy a thread. Threads are
			// subject to garbage collection, like any Lua object. 
			return ctx
		}
	}
	return ctx
}

func toContinuous(values []interface{}) (ctx *Continuous, err error) {

	if 2 <= len(values) && nil != values[1] {
		err = values[1].(error)
	}

	if nil != values[0] {
		var ok bool = false
		ctx, ok = values[0].(*Continuous)
		if !ok {
			if nil != err {
				err = commons.NewTwinceError(err, fmt.Errorf("oooooooo! It is not a Continuous - %v", values[0]))
			} else {
				err = fmt.Errorf("oooooooo! It is not a Continuous - %v", values[0])
			}
		}
	} else if nil == err {
		err = errors.New("oooooooo! return a nil")
	}
	return
}

func (self *LuaDriver) newContinuous(action string, params map[string]string) *Continuous {
	if nil == self.LS {
		return &Continuous{status: LUA_EXECUTE_FAILED,
			Error: errors.New("lua status is nil.")}
	}

	method := &NativeMethod{
		Name:  "get",
		Write: writeActionArguments}

	ctx := &Continuous{
		LS:          self.LS,
		status:      LUA_EXECUTE_END,
		StringValue: action,
		Params:      params,
		on_end:      readActionResult,
		method:      method}

	ctx = self.eval(ctx)
	switch ctx.status {
	case LUA_EXECUTE_CONTINUE:
		ctx.status = LUA_EXECUTE_FAILED
		ctx.Error = errors.New("synchronization call is prohibited while the process of creating thread.")
		return ctx
	case LUA_EXECUTE_END:
		ctx.status = LUA_EXECUTE_FAILED
		ctx.Error = errors.New("'core.lua' is directly exited.")
		return ctx
	case LUA_EXECUTE_YIELD:
		new_th, err := toThread(self.LS, -1)
		if nil != err {
			ctx.status = LUA_EXECUTE_FAILED
			ctx.Error = errors.New("main fiber return error by yeild, " + err.Error())
			return ctx
		}

		ctx.LS = new_th
		ctx.method = method
		ctx.method.Write = nil
		ctx = self.eval(ctx)
		if LUA_EXECUTE_CONTINUE == ctx.status {
			self.waitG.Add(1)
		}
		return ctx
	default:
		ctx.status = LUA_EXECUTE_FAILED
		if nil == ctx.Error {
			ctx.Error = errors.New("switch to main fiber failed.")
		} else {
			ctx.Error = errors.New("switch to main fiber failed, " + ctx.Error.Error())
		}
		return ctx
	}
	return ctx
}

func (driver *LuaDriver) invoke(action string, params map[string]string) (interface{}, error) {
	t := 5 * time.Minute
	old := time.Now()

	values := driver.SafelyCall(t, func() *Continuous {
		return driver.newContinuous(action, params)
	})
	ctx, err := toContinuous(values)
	if nil != err {
		if nil != ctx && LUA_EXECUTE_CONTINUE == ctx.status {
			driver.waitG.Done()
		}
		return nil, err
	}

	if LUA_EXECUTE_CONTINUE == ctx.status {
		defer func() {
			driver.waitG.Done()
		}()

		for LUA_EXECUTE_CONTINUE == ctx.status {
			if nil != ctx.method && nil != ctx.method.Callback {
				ctx.method.Callback(driver, ctx)
			}

			seconds := (time.Now().Second() - old.Second())
			t -= (time.Duration(seconds) * time.Second)
			values := driver.SafelyCall(t, func() *Continuous {
				return driver.eval(ctx)
			})

			ctx, err = toContinuous(values)
			if nil != err {
				return nil, err
			}
		}
	}

	if LUA_EXECUTE_END == ctx.status {
		return ctx.Any, ctx.Error
	}
	return nil, ctx.Error
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
