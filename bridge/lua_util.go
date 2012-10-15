package main

// #include <stdlib.h>
// #include "lua.h"
// #include "lualib.h"
// #include "lauxlib.h"
import "C"
import (
	"errors"
	"fmt"
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
		panic("not implemented")
	case C.LUA_TNUMBER:
		v := C.lua_tonumberx(ls, index, nil)
		return float64(v)
	case C.LUA_TSTRING:
		return toString(ls, index)
	case C.LUA_TTABLE:
		return toTable(ls, index)
	case C.LUA_TFUNCTION:
		panic("convert function is not implemented")
	case C.LUA_TUSERDATA:
		panic("convert userdata is not implemented")
	case C.LUA_TTHREAD:
		panic("convert thread is not implemented")
	default:
		panic("not implemented")
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
			panic("[array and map] is unsupported type")
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
	if nil == any {
		C.lua_pushnil(ls)
		return
	}
	C.lua_pushnil(ls)
}

func pushString(ls *C.lua_State, s string) {
	if "" == s {
		C.lua_pushnil(ls)
		return
	}

	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))

	C.lua_pushstring(ls, cs)
}

func pushParams(ls *C.lua_State, params map[string]string) {
	if nil == params {
		C.lua_pushnil(ls)
		return
	}
	C.lua_createtable(ls, 0, 0)
	for k, v := range params {
		cs := C.CString(k)
		defer C.free(unsafe.Pointer(cs))

		pushString(ls, v)
		C.lua_setfield(ls, -2, cs)
	}
}
