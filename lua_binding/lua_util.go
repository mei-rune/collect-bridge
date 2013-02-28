package lua_binding

// #include <stdlib.h>
// #include "lua.h"
// #include "lualib.h"
// #include "lauxlib.h"
import "C"
import (
	"commons"
	"commons/errutils"
	"fmt"
	"log"
	"os"
	"reflect"
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

func ToString(drv *LuaDriver, index int) (string, commons.RuntimeError) {
	if nil == drv.LS {
		return "", errutils.InternalError("lua_State is nil")
	}

	if 0 == C.lua_isstring(drv.LS, C.int(index)) {
		return "", errutils.InternalError(fmt.Sprintf("stack[%d] is not a string.", index))
	}

	var length C.size_t
	cs := C.lua_tolstring(drv.LS, C.int(index), &length)
	if nil == cs {
		return "", errutils.InternalError("lua_State is not string?")
	}
	return C.GoStringN(cs, C.int(length)), nil
}

func ToAny(drv *LuaDriver, index int) (interface{}, commons.RuntimeError) {
	return toAny(drv.LS, C.int(index))
}

func PushAny(drv *LuaDriver, any interface{}) {
	pushAny(drv.LS, any)
}

func PushParams(drv *LuaDriver, params map[string]string) {
	pushParams(drv.LS, params)
}

func PushString(drv *LuaDriver, s string) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	C.lua_pushstring(drv.LS, cs)
}

func ResumeLuaFiber(drv *LuaDriver, argc int) {
	ret := C.lua_resume(drv.LS, nil, C.int(argc))
	if C.LUA_YIELD != ret {
		err := getError(drv.LS, ret, "test push params failed")
		log.Panicf(err.Error())
	}
}

func pushAnyTest(drv *LuaDriver, any interface{}) interface{} {
	pushString(drv.LS, "test")
	pushAny(drv.LS, any)
	ret := C.lua_resume(drv.LS, nil, 2)
	if C.LUA_YIELD != ret {
		err := getError(drv.LS, ret, "test push any failed")
		log.Panicf(err.Error())
	}
	any, err := toAny(drv.LS, 2)
	if nil != err {
		log.Panicf(err.Error())
	}
	return any
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

// func declareAny(L *C.lua_State) {
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
//}

func toThread(ls *C.lua_State, idx C.int) (*C.lua_State, commons.RuntimeError) {
	if C.LUA_TTHREAD != C.lua_type(ls, idx) {
		return nil, errutils.InternalError("it is not a 'lua_State'")
	}

	new_th := C.lua_tothread(ls, idx)
	if nil == new_th {
		return nil, errutils.InternalError("it is nil")
	}
	return new_th, nil
}

func toAny(ls *C.lua_State, index C.int) (interface{}, commons.RuntimeError) {
	if nil == ls {
		return nil, errutils.InternalError("lua_State is nil")
	}

	t := C.lua_type(ls, index)
	switch t {
	case C.LUA_TNONE:
		return nil, nil
	case C.LUA_TNIL:
		return nil, nil
	case C.LUA_TBOOLEAN:
		if 0 != C.lua_toboolean(ls, index) {
			return true, nil
		}
		return false, nil
	case C.LUA_TLIGHTUSERDATA:
		return nil, errutils.InternalError("convert lightuserdata is not implemented")
	case C.LUA_TNUMBER:
		if iv := C.lua_tointegerx(ls, index, nil); 0 != int64(iv) {
			return int64(iv), nil
		} else if uv := C.lua_tounsignedx(ls, index, nil); 0 != int64(uv) {
			return int64(uv), nil
		} else {
			nu := C.lua_tonumberx(ls, index, nil)
			return float64(nu), nil
		}
	case C.LUA_TSTRING:
		return toString(ls, index)
	case C.LUA_TTABLE:
		return toTable(ls, index)
	case C.LUA_TFUNCTION:
		return nil, errutils.InternalError("convert function is not implemented")
	case C.LUA_TUSERDATA:
		return nil, errutils.InternalError("convert userdata is not implemented")
	case C.LUA_TTHREAD:
		return toThread(ls, index)
	default:
		return nil, errutils.InternalError("not implemented")
	}
	return nil, nil
}

func convertMapToArray(m map[int]interface{}) ([]interface{}, commons.RuntimeError) {
	res := make([]interface{}, 0, len(m)+16)
	for k, v := range m {
		if len(res) > k {
			res[k] = v
		} else if len(res) == k {
			res = append(res, v)
		} else {
			if k > 50000 {
				return nil, errutils.InternalError("ooooooooooooooo! array is too big!")
			}

			for i := len(res); i < k; i++ {
				res = append(res, nil)
			}
			res = append(res, v)
		}
	}
	return res, nil
}
func toTable(ls *C.lua_State, index C.int) (interface{}, commons.RuntimeError) {

	if nil == ls {
		return nil, errutils.InternalError("lua_State is nil")
	}

	if LUA_TTABLE != C.lua_type(ls, index) {
		return nil, errutils.InternalError(fmt.Sprintf("stack[%d] is not a table.", index))
	}

	res1 := make(map[int]interface{}, 10)
	res2 := make(map[string]interface{}, 10)

	if 0 > index {
		index = C.lua_gettop(ls) + index + C.int(1)
	}
	//fmt.Printf("push nil at %d\n", int(index))

	C.lua_pushnil(ls) /* first key */
	for 0 != C.lua_next(ls, index) {
		//fmt.Printf("%s - %s\n",
		//	C.GoString(C.lua_typename(ls, C.lua_type(ls, -2))),
		//	C.GoString(C.lua_typename(ls, C.lua_type(ls, -1))))
		/* 'key' is at index -2 and 'value' at index -1 */
		if 0 != C.lua_isnumber(ls, -2) {
			idx := int(C.lua_tointegerx(ls, -2, nil))
			if 0 == idx {
				return nil, errutils.InternalError(fmt.Sprintf("read index from stack[%d] fail.", index))
			}
			any, err := toAny(ls, -1)
			if nil != err {
				return nil, err
			}
			res1[idx-1] = any
		} else if 0 != C.lua_isstring(ls, -2) {

			any, err := toAny(ls, -1)
			if nil != err {
				return nil, err
			}
			k, err := toString(ls, -2)
			if nil != err {
				return nil, errutils.InternalError("read key failed, " + err.Error())
			}

			res2[k] = any
		} else {
			return nil, errutils.InternalError(fmt.Sprintf("key must is a string or number while read table from stack[%d] fail.", index))
		}
		/* removes 'value'; keeps 'key' for next iteration */
		C.lua_settop(ls, -2) // C.lua_pop(ls, 1)
	}

	if 0 != len(res1) {
		if 0 != len(res2) {
			return nil, errutils.InternalError(fmt.Sprintf("data of stack[%d] is mixed with array and map, it is unsupported type", index))
		}

		return convertMapToArray(res1)
	}

	if 0 == len(res2) {
		return nil, nil
	}

	return res2, nil
}

func toParams(ls *C.lua_State, index C.int) (map[string]string, commons.RuntimeError) {

	if nil == ls {
		return nil, errutils.InternalError("lua_State is nil")
	}

	if LUA_TTABLE != C.lua_type(ls, index) {
		return nil, errutils.InternalError(fmt.Sprintf("stack[%d] is not a table.", index))
	}

	res := make(map[string]string, 10)

	if 0 > index {
		index = C.lua_gettop(ls) + index + C.int(1)
	}

	C.lua_pushnil(ls) /* first key */
	for 0 != C.lua_next(ls, index) {
		/* 'key' is at index -2 and 'value' at index -1 */
		k, err := toString(ls, -2)
		if nil != err {
			return nil, errutils.InternalError("read key failed, " + err.Error())
		}
		v, err := toString(ls, -1)
		if nil != err {
			return nil, errutils.InternalError("read value failed, " + err.Error())
		}

		res[k] = v

		/* removes 'value'; keeps 'key' for next iteration */
		C.lua_settop(ls, -2) // C.lua_pop(ls, 1)
	}
	return res, nil
}

func toInteger(ls *C.lua_State, index C.int) (int, commons.RuntimeError) {
	if nil == ls {
		return 0, errutils.InternalError("lua_State is nil")
	}
	var isnum C.int = 0
	iv := C.lua_tointegerx(ls, index, &isnum)
	if 0 == isnum {
		return 0, errutils.InternalError("It is not a number")
	}
	return int(iv), nil
}

func toString(ls *C.lua_State, index C.int) (string, commons.RuntimeError) {
	if nil == ls {
		return "", errutils.InternalError("lua_State is nil")
	}

	if 0 == C.lua_isstring(ls, index) {
		return "", errutils.InternalError(fmt.Sprintf("stack[%d] is not a string.", index))
	}

	var length C.size_t
	cs := C.lua_tolstring(ls, index, &length)
	if nil == cs {
		return "", errutils.InternalError("lua_State is not string?")
	}
	return C.GoStringN(cs, C.int(length)), nil
}

func toError(ls *C.lua_State, index C.int) commons.RuntimeError {
	if nil == ls {
		return errutils.InternalError("lua_State is nil")
	}
	var length C.size_t
	cs := C.lua_tolstring(ls, index, &length)
	if nil == cs {
		return nil
	}
	return errutils.InternalError(C.GoStringN(cs, C.int(length)))
}

func getError(ls *C.lua_State, ret C.int, msg string) commons.RuntimeError {
	if nil == ls {
		return errutils.InternalError("lua_State is nil")
	}

	var length C.size_t
	cs := C.lua_tolstring(ls, -1, &length)
	if nil == cs {
		return errutils.InternalError(fmt.Sprintf("%s, return code is %d", msg, ret))
	}
	s := C.GoStringN(cs, C.int(length))
	return errutils.InternalError(fmt.Sprintf("%s, commons.RuntimeError message: %s", msg, s))
}

func pushError(ls *C.lua_State, e commons.RuntimeError) {
	if nil == ls {
		return
	}

	if nil == e {
		C.lua_pushnil(ls)
		return
	}

	cs := C.CString(e.Error())
	defer C.free(unsafe.Pointer(cs))

	C.lua_pushstring(ls, cs)
}
func pushAny(ls *C.lua_State, any interface{}) {
	if nil == ls {
		return
	}

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
	case error:
		pushString(ls, v.Error())
	case commons.RuntimeError:
		pushError(ls, v)
	case []interface{}:
		pushArray(ls, v)
	case map[string]interface{}:
		pushMap(ls, v)
	case commons.Result:
		pushMap(ls, map[string]interface{}(v))
	default:
		val := reflect.ValueOf(any)
		switch val.Kind() {
		case reflect.Slice:
			fallthrough
		case reflect.Array:
			C.lua_createtable(ls, 10, 10)
			for i, l := 0, val.Len(); i < l; i++ {
				pushAny(ls, val.Index(i).Interface())
				C.lua_rawseti(ls, -2, C.int(i+1))
			}
		//case reflect.Struct:
		//case reflect.Chan:
		//case reflect.Func:
		//case reflect.Interface:
		// case reflect.Map:

		// 	C.lua_createtable(ls, 0, 0)
		// 	for _, k := range val.MapKeys() {
		// 		cs := C.CString(k)
		// 		pushAny(ls, v)
		// 		C.lua_setfield(ls, -2, cs)

		// 		C.free(unsafe.Pointer(cs))
		// 		cs = nil

		// 	}

		case reflect.Ptr:
			pushAny(ls, val.Interface())
		default:
			log.Panicf("unsupported type - (%T) %v", any, any)
		}
	}
}

func pushMap(ls *C.lua_State, params map[string]interface{}) {
	if nil == ls {
		return
	}

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

	C.lua_createtable(ls, 10, 10)
	for k, v := range params {
		cs := C.CString(k)
		pushAny(ls, v)
		C.lua_setfield(ls, -2, cs)

		C.free(unsafe.Pointer(cs))
		cs = nil
	}
}

func pushArray(ls *C.lua_State, params []interface{}) {
	if nil == ls {
		return
	}

	if nil == params {
		C.lua_pushnil(ls)
		return
	}

	C.lua_createtable(ls, 10, 10)
	for k, v := range params {
		pushAny(ls, v)
		C.lua_rawseti(ls, -2, C.int(k+1))
	}
}

func pushString(ls *C.lua_State, s string) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	C.lua_pushstring(ls, cs)
}

func pushParams(ls *C.lua_State, params map[string]string) {
	if nil == ls {
		return
	}

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

	C.lua_createtable(ls, 10, 10)
	for k, v := range params {
		cs = C.CString(k)
		pushString(ls, v)
		C.lua_setfield(ls, -2, cs)

		C.free(unsafe.Pointer(cs))
		cs = nil
	}
}
