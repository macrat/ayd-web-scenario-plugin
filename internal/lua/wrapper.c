#include <lua.h>
#include <lauxlib.h>

int f_lua_upvalueindex(int index) {
	return lua_upvalueindex(index);
}

int f_luaL_getmetatable(lua_State* L, const char* tname) {
    return luaL_getmetatable(L, tname);
}

int error_message_handler(lua_State* L) {
    if (lua_gettop(L) > 0 && lua_type(L, -1) == LUA_TSTRING) {
        luaL_traceback(L, L, lua_tostring(L, -1), 2);
    }
    return 1;
}
