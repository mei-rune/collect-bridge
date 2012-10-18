package main

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
	"unsafe"
)

const (
	LUA_TTABLE = C.LUA_TTABLE
)

func fileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func pushAnyTest(drv *LuaDriver, any interface{}) interface{} {
	pushString(drv.ls, "test")
	pushAny(drv.ls, any)
	ret := C.lua_resume(drv.ls, nil, 2)
	if C.LUA_YIELD != ret {
		err := getError(drv.ls, ret, "test push any failed")
		log.Panicf(err.Error())
	}
	return toAny(drv.ls, 2)
}

// unsigned int* go_check_any(lua_State* L, int index)
// {
// 	unsigned int* fid = (unsigned int*)luaL_checkudata(L,index,"Go.Any");
// 	luaL_argcheck(L, NULL != fid, index, "'Go.Any' expected");
// 	return fid;
// }

// //wrapper for callgofunction
// int go_to_number(lua_State* L)
// {
// 	unsigned int *fid = go_check_any(L,1);

// 	lua_remove(L,1);
// 	return golua_callgofunction(*gi,*fid);
// }

func declareAny(L *C.lua_State) {
	// /* create the GoLua.GoFunction metatable */
	// C.luaL_newmetatable(L, "Go.Any")
	// //pushkey
	// C.lua_pushliteral(L, "__call")
	// //push value
	// C.lua_pushcfunction(L, &callback_function)
	// //t[__call] = &callback_function
	// C.lua_settable(L, -3)
	// //push key
	// C.lua_pushliteral(L, "__gc")
	// //pushvalue
	// C.lua_pushcfunction(L, &gchook_wrapper)
	// C.lua_settable(L, -3)
	// C.lua_pop(L, 1)
}

func toAny(ls *C.lua_State, index C.int) interface{} {
	t := C.lua_type(ls, index)
	switch t {
	case C.LUA_TNONE:
		return nil
	case C.LUA_TNIL:
		return nil
	case C.LUA_TBOOLEAN:
		if 0 != C.lua_toboolean(ls, index) {
			return true
		}
		return false
	case C.LUA_TLIGHTUSERDATA:
		log.Panicf("not implemented")
	case C.LUA_TNUMBER:
		if iv := C.lua_tointegerx(ls, index, nil); 0 != int64(iv) {
			return int64(iv)
		} else if uv := C.lua_tounsignedx(ls, index, nil); 0 != int64(uv) {
			return int64(uv)
		} else {
			nu := C.lua_tonumberx(ls, index, nil)
			return float64(nu)
		}
	case C.LUA_TSTRING:
		return toString(ls, index)
	case C.LUA_TTABLE:
		return toTable(ls, index)
	case C.LUA_TFUNCTION:
		log.Panicf("convert function is not implemented")
	case C.LUA_TUSERDATA:
		log.Panicf("convert userdata is not implemented")
	case C.LUA_TTHREAD:
		log.Panicf("convert thread is not implemented")
	default:
		log.Panicf("not implemented")
	}
	return nil
}

func toTable(ls *C.lua_State, index C.int) interface{} {

	if LUA_TTABLE == C.lua_type(ls, index) {
		return nil
	}
	res1 := make(map[int]interface{})
	res2 := make(map[string]interface{})

	C.lua_pushnil(ls) /* first key */
	for 0 != C.lua_next(ls, index) {
		/* 'key' is at index -2 and 'value' at index -1 */
		if 0 == C.lua_isnumber(ls, -2) {
			res1[int(C.lua_tointegerx(ls, -2, nil))] = toAny(ls, -1)
		} else {
			res2[toString(ls, -2)] = toAny(ls, -1)
		}
		/* removes 'value'; keeps 'key' for next iteration */
		C.lua_settop(ls, -2) // C.lua_pop(ls, 1)
	}

	C.lua_settop(ls, -2) // C.lua_pop(ls, 1) /* removes 'key' */
	if 0 != len(res1) {
		if 0 != len(res2) {
			log.Panicf("data is mixed with array and map, it is unsupported type.")
		}
		return res1
	}

	if 0 == len(res2) {
		return nil
	}

	return res2
}

func toParams(ls *C.lua_State, index C.int) map[string]string {

	if LUA_TTABLE == C.lua_type(ls, index) {
		return nil
	}
	res := make(map[string]string)

	C.lua_pushnil(ls) /* first key */
	for 0 != C.lua_next(ls, index) {
		/* 'key' is at index -2 and 'value' at index -1 */
		res[toString(ls, -2)] = toString(ls, -2)

		/* removes 'value'; keeps 'key' for next iteration */
		C.lua_settop(ls, -2) // C.lua_pop(ls, 1)
	}

	C.lua_settop(ls, -2) // C.lua_pop(ls, 1) /* removes 'key' */
	return res
}

func toString(ls *C.lua_State, index C.int) string {
	var length C.size_t
	cs := C.lua_tolstring(ls, index, &length)
	if nil == cs {
		return ""
	}
	return C.GoStringN(cs, C.int(length))
}

func toError(ls *C.lua_State, index C.int) error {
	var length C.size_t
	cs := C.lua_tolstring(ls, index, &length)
	if nil == cs {
		return nil
	}
	return errors.New(C.GoStringN(cs, C.int(length)))
}

func getError(ls *C.lua_State, ret C.int, msg string) error {
	var length C.size_t
	cs := C.lua_tolstring(ls, -1, &length)
	if nil == cs {
		return fmt.Errorf("%s, return code is %d", msg, ret)
	}
	s := C.GoStringN(cs, C.int(length))
	return fmt.Errorf("%s, error message: %s", msg, s)
}

func pushError(ls *C.lua_State, e error) {
	if nil == e {
		C.lua_pushnil(ls)
		return
	}

	cs := C.CString(e.Error())
	defer C.free(unsafe.Pointer(cs))

	C.lua_pushstring(ls, cs)
}
func pushAny(ls *C.lua_State, any interface{}) {
	switch v := any.(type) {
	case nil:
		C.lua_pushnil(ls)
	case bool:
		if v {
			C.lua_pushboolean(ls, 1)
		} else {
			C.lua_pushboolean(ls, 0)
		}
	case uint:
		C.lua_pushunsigned(ls, C.lua_Unsigned(v))
	case uint8:
		C.lua_pushunsigned(ls, C.lua_Unsigned(v))
	case uint16:
		C.lua_pushunsigned(ls, C.lua_Unsigned(v))
	case uint32:
		C.lua_pushunsigned(ls, C.lua_Unsigned(v))
	case uint64:
		C.lua_pushnumber(ls, C.lua_Number(v))
	case int:
		C.lua_pushinteger(ls, C.lua_Integer(v))
	case int8:
		C.lua_pushinteger(ls, C.lua_Integer(v))
	case int16:
		C.lua_pushinteger(ls, C.lua_Integer(v))
	case int32:
		C.lua_pushinteger(ls, C.lua_Integer(v))
	case int64:
		C.lua_pushnumber(ls, C.lua_Number(v))
	case float32:
		C.lua_pushnumber(ls, C.lua_Number(v))
	case float64:
		C.lua_pushnumber(ls, C.lua_Number(v))
	case string:
		pushString(ls, v)
	case []interface{}:
		pushArray(ls, v)
	case map[string]interface{}:
		pushMap(ls, v)
	default:
		log.Panicf("unsupported type - %v", any)
	}
}

func pushMap(ls *C.lua_State, params map[string]interface{}) {
	if nil == params {
		C.lua_pushnil(ls)
		return
	}
	var cs *C.char

	defer func() {
		if nil != cs {
			C.free(unsafe.Pointer(cs))
		}
	}()

	C.lua_createtable(ls, 0, 0)
	for k, v := range params {
		cs := C.CString(k)
		pushAny(ls, v)
		C.lua_setfield(ls, -2, cs)

		C.free(unsafe.Pointer(cs))
		cs = nil
	}
}

func pushArray(ls *C.lua_State, params []interface{}) {
	if nil == params {
		C.lua_pushnil(ls)
		return
	}

	C.lua_createtable(ls, 0, 0)
	for k, v := range params {
		pushAny(ls, v)
		C.lua_rawseti(ls, -2, C.int(k))
	}
}

func pushString(ls *C.lua_State, s string) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	C.lua_pushstring(ls, cs)
}

func pushParams(ls *C.lua_State, params map[string]string) {
	if nil == params {
		C.lua_pushnil(ls)
		return
	}
	var cs *C.char

	defer func() {
		if nil != cs {
			C.free(unsafe.Pointer(cs))
		}
	}()

	C.lua_createtable(ls, 0, 0)
	for k, v := range params {
		cs = C.CString(k)
		pushString(ls, v)
		C.lua_setfield(ls, -2, cs)

		C.free(unsafe.Pointer(cs))
		cs = nil
	}
}
