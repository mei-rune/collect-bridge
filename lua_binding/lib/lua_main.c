 
#include <stdlib.h>
#include "lua.h"
#include "lualib.h"
#include "lauxlib.h"


void printError(lua_State* ls, int ret, const char* msg) {
    size_t length = 0;
    char * s;
    int ret2;
    ret2 = lua_gettop (ls);
    if (0 == ret2) {
        printf("%s, return code is %d\n", msg, ret);
        return;
    }

    const char* cs = luaL_checklstring(ls, -1, &length);
    if (0 == cs) {
        printf("%s, return code is %d\n", msg, ret);
        return;
    }
    s = (char*)malloc(length + 10);
    memcpy(s, cs, length);
    s[length] = 0;
    printf("%s, error message: %s\n", msg, s);
}

lua_State* start() {
    int ret;
    lua_State* ls;
    
    ls = luaL_newstate();
    luaL_openlibs(ls);
    ret = luaL_loadfilex(ls, "lua_init.lua", 0);
    if (LUA_ERRFILE == ret) {
        printf("'lua_init.lua' read fail\n");
        return 0;
    } else if (0 != ret) {
        printError(ls, ret, "load 'lua_init.lua' failed");
        return 0;
    }

    ret = lua_resume(ls, 0, 0);
    if (LUA_YIELD != ret) {
        printError(ls, ret, "launch main fiber failed");
        return 0;
    }
    return ls;
}

void stop(lua_State* ls) {
    int ret;

    if (0 == ls) {
        return ;
    }

    ret = lua_status(ls);
    if (LUA_YIELD != ret) {
        printError(ls, ret, "stop main fiber failed");
        return 0;
    }

    lua_pushstring(ls, "__exit__");

    ret = lua_resume(ls, 0, 1);
    if (0 != ret) {
        printError(ls, ret, "stop main fiber failed");
        return ;
    }

    lua_close(ls);
}

void run(lua_State* ls, const char* action) {
    int ret;
    lua_State* new_th;

    lua_pushstring(ls, action);
    lua_pushnil(ls);

    ret = lua_resume(ls, 0, 2);
    if (LUA_YIELD != ret) {
        if (0 == ret) {
            printf("'lua_init.lua' is directly exited.\n");
            return 0;
        } else {
            printError(ls, ret, "switch to main fiber failed");
            return 0;
        }
    }

    if (LUA_TTHREAD != lua_type(ls, -1)) {
        printf("main fiber return value by yeild is not lua_Status type\n");
        return 0;
    }

    new_th = lua_tothread(ls, -1);
    if (0 == new_th) {
        printf("main fiber return value by yeild is nil\n");
        return 0;
    }


    ret =lua_resume(new_th, ls, 0);
    switch (ret) {
    case 0:
        printError(new_th, -1, "value is ");
        break;
    default:
        printError(new_th, ret, "run failed");
        break ;
    }
//end:
//    lua_close(new_th);
}


int main(int argc, char* argv[]) {
    lua_State* ls = start();
    if (0 == ls) {
        return -1;
    }


    run(ls, "get");
    stop(ls);
    return 0;
}