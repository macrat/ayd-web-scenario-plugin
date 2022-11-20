#include <lua.h>
#include <lauxlib.h>

int f_lua_upvalueindex(int index) {
	return lua_upvalueindex(index);
}

int f_luaL_getmetatable(lua_State* L, const char* tname) {
    return luaL_getmetatable(L, tname);
}
