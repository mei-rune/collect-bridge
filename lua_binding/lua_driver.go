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
	"commons/as"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"snmp"
	"sync"
	"time"
	"unsafe"
)

const (
	lua_init_script string = `
local mj = {}

mj.DEBUG = 9000
mj.INFO = 6000
mj.WARN = 4000
mj.ERROR = 2000
mj.FATAL = 1000
mj.SYSTEM = 0

function mj.receive ()
    local action, params = coroutine.yield()
    return action, params
end

function mj.send_and_recv ( ...)
    local action, params = coroutine.yield( ...)
    return action, params
end

function mj.log(level, msg)
  if "number" ~= type(level) then
    return nil, "'params' is not a table."
  end

  coroutine.yield("log", level, msg)
end

function mj.execute(schema, action, params)
  if "table" ~= type(params) then
    return nil, "'params' is not a table."
  end
  return coroutine.yield(action, schema, params)
end

function mj.execute_module(module_name, action, params)
  module = require(module_name)
  if nil == module then
    return nil, "module '"..module_name.."' is not exists."
  end
  func = module[action]
  if nil == func then
    return nil, "method '"..action.."' is not implemented in module '"..module_name.."'."
  end

  return func(params)
end

function mj.execute_script(action, script, params)
  if 'string' ~= type(script) then
    return nil, "'script' is not a string."
  end
  local env = {["mj"] = mj,
   ["action"] = action,
   ['params'] = params}
  setmetatable(env, _ENV)
  func = assert(load(script, nil, 'bt', env))
  return func()
end

function mj.execute_task(action, params)
  --if nil == task then
  --  print("params = nil")
  --end

  return coroutine.create(function()
      if nil == params then
        return nil, "'params' is nil."
      end
      if "table" ~= type(params) then
        return nil, "'params' is not a table, actual is '"..type(params).."'." 
      end
      schema = params["schema"]
      if nil == schema then
        return nil, "'schema' is nil"
      elseif "script" == schema then
        return mj.execute_script(action, params["script"], params)
      else
        return mj.execute_module(schema, action, params)
      end
    end)
end


function mj.loop()

  mj.os = __mj_os or "unknown"  -- 386, amd64, or arm.
  mj.arch = __mj_arch or "unknown" -- darwin, freebsd, linux or windows
  mj.execute_directory = __mj_execute_directory or "."
  mj.work_directory = __mj_work_directory or "."

  local ext = ".so"
  local sep = "/"
  if mj.os == "windows" then
    ext = ".dll"
    sep = "\\"
  end

  if nil ~= __mj_execute_directory then
    package.path = package.path .. ";" .. mj.execute_directory .. sep .. "modules"..sep.."?.lua" ..
       ";" .. mj.execute_directory .. sep .. "modules" .. sep .. "?" .. sep .. "init.lua"

    package.cpath = package.cpath .. ";" .. mj.execute_directory .. sep .."modules" .. sep .. "?" .. ext ..
        ";" .. mj.execute_directory .. sep .. "modules" .. sep .. "?" .. sep .. "loadall" .. ext
  end

  if nil ~= __mj_work_directory then
    package.path = package.path .. ";" .. mj.work_directory .. sep .. "modules" .. sep .. "?.lua" ..
       ";" .. mj.work_directory .. sep .. "modules" .. sep .. "?" .. sep .. "init.lua"

    package.cpath = package.cpath .. ";" .. mj.work_directory .. sep .. "modules" .. sep .. "?" .. ext ..
        ";" .. mj.work_directory .. sep .. "modules" .. sep .. "?" .. sep .. "loadall" .. ext
  end


  mj.log(SYSTEM, "lua enter looping")
  local action, params = mj.receive()  -- get new value
  while "__exit__" ~= action do
    mj.log(SYSTEM, "lua vm receive - '"..action.."'")

    co = mj.execute_task(action, params)
    action, params = mj.send_and_recv(co)
  end
  mj.log(SYSTEM, "lua exit looping")
end

_G["mj"] = mj
package.loaded["mj"] = mj
package.preload["mj"] = mj
mj.log(SYSTEM, "welcome to lua vm")
mj.loop ()
`

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

	method_missing = &NativeMethod{
		Name:     "method_missing",
		Read:     nil,
		Write:    writeStdArguments,
		Callback: nil}
)

type LuaDriver struct {
	snmp.Svc
	init_path string
	LS        *C.lua_State
	waitG     sync.WaitGroup

	methods        map[string]*NativeMethod
	method_missing *NativeMethod
}

type Continuous struct {
	LS     *C.lua_State
	status LUA_CODE
	action string
	method *NativeMethod

	unshift func(drv *LuaDriver, ctx *Continuous)
	push    func(drv *LuaDriver, ctx *Continuous) (int, error)

	Error       error
	IntValue    int
	StringValue string
	Params      map[string]string
	Any         interface{}
}

func readStdArguments(drv *LuaDriver, ctx *Continuous) {
	ctx.StringValue, ctx.Error = drv.ToStringParam(2)
	if nil != ctx.Error {
		return
	}
	ctx.Params, ctx.Error = drv.ToParamsParam(3)
}

func writeStdArguments(drv *LuaDriver, ctx *Continuous) (int, error) {
	err := drv.PushAnyParam(ctx.Any)
	if nil != err {
		return -1, err
	}
	err = drv.PushErrorParam(ctx.Error)
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
	driver.CallbackWith(&NativeMethod{
		Name:  "get",
		Read:  readStdArguments,
		Write: writeStdArguments,
		Callback: func(drv *LuaDriver, ctx *Continuous) {
			drv, ok := commons.Connect(ctx.StringValue)
			if !ok {
				drv.Error = fmt.Errorf("driver '%s' is not exists.", ctx.StringValue)
				return
			}

			ctx.Any, ctx.Error = drv.Get(ctx.Params)
		}}, &NativeMethod{
		Name:  "put",
		Read:  readStdArguments,
		Write: writeStdArguments,
		Callback: func(drv *LuaDriver, ctx *Continuous) {
			drv, ok := commons.Connect(ctx.StringValue)
			if !ok {
				drv.Error = fmt.Errorf("driver '%s' is not exists.", ctx.StringValue)
				return
			}

			ctx.Any, ctx.Error = drv.Put(ctx.Params)
			return nil
		}}, &NativeMethod{
		Name:  "create",
		Read:  readStdArguments,
		Write: writeStdArguments,
		Callback: func(drv *LuaDriver, ctx *Continuous) {
			drv, ok := commons.Connect(ctx.StringValue)
			if !ok {
				drv.Error = fmt.Errorf("driver '%s' is not exists.", ctx.StringValue)
				return
			}

			ctx.Any, ctx.Error = drv.Create(ctx.Params)
			return nil
		}}, &NativeMethod{
		Name:  "delete",
		Read:  readStdArguments,
		Write: writeStdArguments,
		Callback: func(drv *LuaDriver, ctx *Continuous) {
			drv, ok := commons.Connect(ctx.StringValue)
			if !ok {
				drv.Error = fmt.Errorf("driver '%s' is not exists.", ctx.StringValue)
				return
			}

			ctx.Any, ctx.Error = drv.Delete(ctx.Params)
			return nil
		}}, &NativeMethod{
		Name: "log",
		Read: func(drv *LuaDriver, ctx *Continuous) {
			ctx.IntValue, _ = ctx.ToIntParam(2)
			ctx.StringValue, _ = ctx.ToStringParam(3)
		},
		Write: func(drv *LuaDriver, ctx *Continuous) (int, error) {
			return 0, nil
		},
		Callback: func(drv *LuaDriver, ctx *Continuous) {
			drv.Logger.Println(ctx.StringValue)
		}})
	return driver
}

func (self *LuaDriver) ToParamsParam(idx int) (map[string]string, error) {
	return toParams(self.LS, C.int(idx)), nil
}

func (self *LuaDriver) ToStringParam(idx int) (string, error) {
	return toString(self.LS, C.int(idx)), nil
}

func (self *LuaDriver) ToIntParam(idx int) (int, error) {
	return toInteger(self.LS, C.int(idx)), nil
}

func (self *LuaDriver) PushAnyParam(any interface{}) error {
	pushAny(self.LS, any)
	return nil
}

func (self *LuaDriver) PushErrorParam(e error) error {
	pushError(self.LS, e)
	return nil
}

func (self *LuaDriver) CallbackWith(methods ...*NativeMethod) error {
	for _, m := range methods {
		if nil == m {
			return nil
		}
		if "" == m.Name {
			return errors.New("'name' is empty.")
		}
		if nil == m.Callback {
			return errors.New("'callback' of '" + m.Name + "' is nil.")
		}
		if _, ok := self.methods[m.Name]; ok {
			return errors.New("'" + m.Name + "' is already exists.")
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
		driver.Logger.Printf("LuaDriver: '%s' is not exist.", pa)
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
		driver.Logger.Panicf("'" + driver.init_path + "' read fail")
	} else if 0 != ret {
		driver.Logger.Panicf(getError(ls, ret, "load '"+driver.init_path+"' failed").Error())
	}

	ctx := &Continuous{method: method_init_lua}

	ctx = driver.eval(ctx)
	for LUA_EXECUTE_CONTINUE == ctx.status {
		if nil != nil != ctx.method && nil != ctx.method.Callback {
			ctx.method.Callback(ctx)
		}
		ctx = driver.eval(ctx)
	}

	if LUA_EXECUTE_YIELD != ctx.status {
		driver.Logger.Panicf(getError(ls, ctx.IntValue, "launch main fiber failed").Error())
	}

	driver.LS = ls
	ls = nil
	driver.Logger.Println("driver is started!")
}

func (driver *LuaDriver) atStop() {
	if nil == driver.LS {
		return
	}

	ret := C.lua_status(driver.LS)
	if C.LUA_YIELD != ret {
		driver.Logger.Panicf(getError(driver.LS, ret, "stop main fiber failed").Error())
	}

	pushString(driver.LS, "__exit__")
	ret = driver.eval(driver.LS, nil, 1)
	if 0 != ret {
		driver.Logger.Panicf(getError(driver.LS, ret, "stop main fiber failed").Error())
	}

	driver.Logger.Println("wait for all fibers to exit!")
	driver.waitG.Wait()
	driver.Logger.Println("all fibers is exited!")

	C.lua_close(driver.LS)
	driver.LS = nil
	driver.Logger.Println("driver is exited!")
}

func (self *LuaDriver) newContinuous(action string, params map[string]string) *Continuous {
	if nil == self.LS {
		return &Continuous{status: LUA_EXECUTE_FAILED,
			err: errors.New("lua status is nil.")}
	}

	ctx := &Continuous{
		LS:      nil,
		status:  LUA_EXECUTE_FAILED,
		unshift: readStdArguments,
		method: &NativeMethod{
			Name:     "get",
			Read:     nil,
			Write:    writeStdArguments,
			Callback: nil}}

	ctx = self.eval(self.LS, ctX)
	if LUA_EXECUTE_YIELD != ctx.status {
		if LUA_EXECUTE_END == ctx.status {
			ctx.status = LUA_EXECUTE_FAILED
			ctx.Error = errors.New("'core.lua' is directly exited.")
			return ctx
		} else {
			ctx.status = LUA_EXECUTE_FAILED
			ctx.Error = errors.New("switch to main fiber failed.")
			return ctx
		}
	}

	if nil == self.LS { // check for muti-thread
		ctx.status = LUA_EXECUTE_FAILED
		ctx.Error = errors.New("lua status is nil, exited?")
		return ctx
	}

	if C.LUA_TTHREAD != C.lua_type(self.LS, -1) {
		ctx.status = LUA_EXECUTE_FAILED
		ctx.Error = errors.New("main fiber return value by yeild is not 'lua_State' type")
		return ctx
	}

	new_th := C.lua_tothread(self.LS, -1)
	if nil == new_th {
		ctx.status = LUA_EXECUTE_FAILED
		ctx.Error = errors.New("main fiber return value by yeild is nil")
		return ctx
	}
	ctx.LS = new_th

	ctx = eval(ctx)
	if LUA_EXECUTE_CONTINUE == ctx.status {
		self.waitG.Add(1)
	}
	return ct
}

func (self *LuaDriver) eval(ct *Continuous) *Continuous {
	var ret int = 0
	var argc int = -1
	var ok bool = false

	for {
		argc = 0
		if nil != ct.method.Write {
			argc, ct.Error = ct.method.Write(self, ct)
			if nil != ct.Error {
				ct.status = LUA_EXECUTE_FAILED
				ct.IntValue = C.LUA_ERRERR
				ct.Error = "push arguments failed - " + ct.Error.Error()
				return ct
			}
		}
		if nil == ct.LS {
			ret = C.lua_resume(driver.LS, nil, argc)
		} else {
			ret = C.lua_resume(ct.LS, driver.LS, argc)
		}

		switch ret {
		case 0:
			ct.status = LUA_EXECUTE_END
			if nil != ct.unshift {
				ct.unshift(ct)
			}
			// There is no explicit function to close or to destroy a thread. Threads are
			// subject to garbage collection, like any Lua object. 
			return ct
		case C.LUA_YIELD:
			if 0 == C.lua_gettop(ls) {
				ct.status = LUA_EXECUTE_YIELD
				ct.IntValue = int(ret)
				ct.Error = errors.New("script execute failed - return arguments is empty.")
				return ct
			}

			if 0 == C.lua_isstring(ls, 1) {
				ct.status = LUA_EXECUTE_FAILED
				ct.IntValue = int(ret)
				ct.Error = errors.New("script execute failed - return first argument is not string.")
				return ct
			}

			action := toString(ls, 1)
			ct.method, ok = self.methods[action]
			if !ok {
				ct.Error = fmt.Errorf("unsupport action '%s'", action)
				ct.method = method_missing
			}

			if nil != ct.method.Read {
				ct.method.Read(self, ct)
			}

			if nil != ct.method.Callback {
				ct.status = LUA_EXECUTE_CONTINUE
				return ct
			}
		default:
			ct.status = LUA_EXECUTE_FAILED
			ct.IntValue = int(ret)
			ct.Error = getError(ct.LS, ret, "script execute failed - ")
			// There is no explicit function to close or to destroy a thread. Threads are
			// subject to garbage collection, like any Lua object. 
			return ct
		}
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
			if "log" == ct.action {
				if nil != driver.Logger {
					driver.Logger.Println(ct.err.Error())
				}
				ct.any = nil
				ct.err = nil
			} else {
				ct.any, ct.err = driver.executeTask(ct.drv, ct.action, ct.params)
			}

			seconds := (time.Now().Second() - old.Second())
			t -= (time.Duration(seconds) * time.Second)
			values := driver.SafelyCall(t, func() *Continuous {
				return driver.againContinue(ct)
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
