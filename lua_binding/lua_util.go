package lua_binding

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

func toThread(ls *C.lua_State, idx C.int) (*C.lua_State, error) {
	if C.LUA_TTHREAD != C.lua_type(ls, idx) {
		return nil, errors.New("it is not a 'lua_State'")
	}

	new_th := C.lua_tothread(ls, idx)
	if nil == new_th {
		return nil, errors.New("it is nil")
	}
	return new_th, nil
}

func toAny(ls *C.lua_State, index C.int) (interface{}, error) {
	if nil == ls {
		return nil, errors.New("lua_State is nil")
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
		return nil, errors.New("convert lightuserdata is not implemented")
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
		return toString(ls, index), nil
	case C.LUA_TTABLE:
		return toTable(ls, index)
	case C.LUA_TFUNCTION:
		return nil, errors.New("convert function is not implemented")
	case C.LUA_TUSERDATA:
		return nil, errors.New("convert userdata is not implemented")
	case C.LUA_TTHREAD:
		return toThread(ls, index)
	default:
		return nil, errors.New("not implemented")
	}
	return nil, nil
}

func convertMapToArray(m map[int]interface{}) ([]interface{}, error) {
	res := make([]interface{}, 0, len(m)+16)
	for k, v := range m {
		if len(res) > k {
			res[k] = v
		} else if len(res) == k {
			res = append(res, v)
		} else {
			if k > 50000 {
				return nil, errors.New("ooooooooooooooo! array is too big!")
			}

			for i := len(res); i < k; i++ {
				res = append(res, nil)
			}
			res = append(res, v)
		}
	}
	return res, nil
}
func toTable(ls *C.lua_State, index C.int) (interface{}, error) {

	if nil == ls {
		return nil, errors.New("lua_State is nil")
	}

	if LUA_TTABLE != C.lua_type(ls, index) {
		return nil, fmt.Errorf("stack[%d] is not a table.", index)
	}

	res1 := make(map[int]interface{})
	res2 := make(map[string]interface{})

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
				return nil, fmt.Errorf("read index from stack[%d] fail.", index)
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
			res2[toString(ls, -2)] = any
		} else {
			return nil, fmt.Errorf("key must is a string or number while read table from stack[%d] fail.", index)
		}
		/* removes 'value'; keeps 'key' for next iteration */
		C.lua_settop(ls, -2) // C.lua_pop(ls, 1)
	}

	if 0 != len(res1) {
		if 0 != len(res2) {
			return nil, fmt.Errorf("data of stack[%d] is mixed with array and map, it is unsupported type", index)
		}

		return convertMapToArray(res1)
	}

	if 0 == len(res2) {
		return nil, nil
	}

	return res2, nil
}

func toParams(ls *C.lua_State, index C.int) map[string]string {
	if nil == ls {
		return nil
	}

	if LUA_TTABLE != C.lua_type(ls, index) {
		return nil
	}

	res := make(map[string]string)

	if 0 > index {
		index = C.lua_gettop(ls) + index + C.int(1)
	}

	C.lua_pushnil(ls) /* first key */
	for 0 != C.lua_next(ls, index) {
		if 0 == C.lua_isstring(ls, -2) {
			log.Panicln("key must is a string.")
		}

		/* 'key' is at index -2 and 'value' at index -1 */
		res[toString(ls, -2)] = toString(ls, -1)

		/* removes 'value'; keeps 'key' for next iteration */
		C.lua_settop(ls, -2) // C.lua_pop(ls, 1)
	}
	return res
}

func toInteger(ls *C.lua_State, index C.int) int {
	if nil == ls {
		return -1
	}
	iv := C.lua_tointegerx(ls, index, nil)
	return int(iv)
}

func toString(ls *C.lua_State, index C.int) string {
	if nil == ls {
		return ""
	}
	var length C.size_t
	cs := C.lua_tolstring(ls, index, &length)
	if nil == cs {
		return ""
	}
	return C.GoStringN(cs, C.int(length))
}

func toError(ls *C.lua_State, index C.int) error {
	if nil == ls {
		return nil
	}
	var length C.size_t
	cs := C.lua_tolstring(ls, index, &length)
	if nil == cs {
		return nil
	}
	return errors.New(C.GoStringN(cs, C.int(length)))
}

func getError(ls *C.lua_State, ret C.int, msg string) error {
	if nil == ls {
		return nil
	}

	var length C.size_t
	cs := C.lua_tolstring(ls, -1, &length)
	if nil == cs {
		return fmt.Errorf("%s, return code is %d", msg, ret)
	}
	s := C.GoStringN(cs, C.int(length))
	return fmt.Errorf("%s, error message: %s", msg, s)
}

func pushError(ls *C.lua_State, e error) {
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
	case []interface{}:
		pushArray(ls, v)
	case map[string]interface{}:
		pushMap(ls, v)
	default:
		log.Panicf("unsupported type - %v", any)
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
	if nil == ls {
		return
	}

	if nil == params {
		C.lua_pushnil(ls)
		return
	}

	C.lua_createtable(ls, 0, 0)
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

	C.lua_createtable(ls, 0, 0)
	for k, v := range params {
		cs = C.CString(k)
		pushString(ls, v)
		C.lua_setfield(ls, -2, cs)

		C.free(unsafe.Pointer(cs))
		cs = nil
	}
}
