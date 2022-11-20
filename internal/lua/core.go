package lua

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"unsafe"
)

/*
#include <stdlib.h>

#cgo CFLAGS: -I./clua
#cgo LDFLAGS: -L./clua -llua -lm -ldl
#include <lua.h>
#include <lualib.h>
#include <lauxlib.h>

extern int gUserdataCollector(lua_State*);
extern int gFunctionWrapper(lua_State*);
extern int gFunctionCollector(lua_State*);
extern void gHookWrapper(lua_State*, lua_Debug*);
extern char* gReader(lua_State*, void*, size_t*);

extern int f_lua_upvalueindex(int);
extern int f_luaL_getmetatable(lua_State*, const char*);
*/
import "C"

var (
	ErrRuntime = errors.New("runtime error")
	ErrMemory  = errors.New("memory allocation error")
	ErrError   = errors.New("error")
	ErrSyntax  = errors.New("syntax error")
	ErrYield   = errors.New("yield")
	ErrFile    = errors.New("file error")
	ErrUnknown = errors.New("unknown error")
)

const (
	MultRet = int(C.LUA_MULTRET)
)

type State struct {
	state *C.lua_State
}

func NewState() (*State, error) {
	s := C.luaL_newstate()
	if s == nil {
		return nil, ErrMemory
	}

	C.luaL_openlibs(s)

	L := &State{s}

	L.PushFunction(pcall)
	L.SetGlobal("pcall")

	L.PushFunction(xpcall)
	L.SetGlobal("xpcall")

	return L, nil
}

func (L *State) Close() error {
	C.lua_close(L.state)
	return nil
}

func (L *State) GetTop() int {
	return int(C.lua_gettop(L.state))
}

func (L *State) SetTop(index int) {
	C.lua_settop(L.state, C.int(index))
}

func (L *State) AbsIndex(index int) int {
	return int(C.lua_absindex(L.state, C.int(index)))
}

func (L *State) Type(index int) Type {
	return Type(C.lua_type(L.state, C.int(index)))
}

func (L *State) IsInteger(index int) bool {
	return C.lua_isinteger(L.state, C.int(index)) != 0
}

func (L *State) ToBoolean(index int) bool {
	return int(C.lua_toboolean(L.state, C.int(index))) != 0
}

func (L *State) ToInteger(index int) int64 {
	return int64(C.lua_tointegerx(L.state, C.int(index), nil))
}

// ToString calls luaL_tolstring.
func (L *State) ToString(index int) string {
	var l C.size_t
	if L.Type(index) == String {
		s := C.lua_tolstring(L.state, C.int(index), &l)
		return C.GoStringN(s, C.int(l))
	} else {
		s := C.luaL_tolstring(L.state, C.int(index), &l)
		str := C.GoStringN(s, C.int(l))
		L.Pop(1)
		return str
	}
}

func (L *State) ToNumber(index int) float64 {
	return float64(C.lua_tonumberx(L.state, C.int(index), nil))
}

var (
	userdata = newStore[any]()
)

// ToUserdata returns value of userdata.
// This method returns nil if the userdata is not set by this library.
func (L *State) ToUserdata(index int) any {
	id := (*C.int)(unsafe.Pointer(C.lua_touserdata(L.state, C.int(index))))
	return userdata.Get(int(*id))
}

func (L *State) PushBoolean(b bool) {
	i := C.int(0)
	if b {
		i = 1
	}
	C.lua_pushboolean(L.state, i)
}

func (L *State) PushInteger(n int64) {
	C.lua_pushinteger(L.state, C.longlong(n))
}

func (L *State) PushString(s string) {
	str := C.CString(s)
	defer C.free(unsafe.Pointer(str))
	C.lua_pushlstring(L.state, str, C.size_t(len(s)))
}

func (L *State) PushNil() {
	C.lua_pushnil(L.state)
}

func (L *State) PushNumber(n float64) {
	C.lua_pushnumber(L.state, C.double(n))
}

//export gUserdataCollector
func gUserdataCollector(state *C.lua_State) C.int {
	id := (*C.int)(unsafe.Pointer(C.lua_touserdata(state, 1)))
	userdata.Pop(int(*id))
	return 0
}

// PushUserdata pushes any value in go to the stack.
// This method sets metatable for garbage collecting. You should reuse that metatable to prevent memory leak if you use your own metatable.
func (L *State) PushUserdata(v any) {
	id := (*C.int)(unsafe.Pointer(C.lua_newuserdatauv(L.state, C.sizeof_int, 1)))
	*id = C.int(userdata.Push(v))

	L.CreateTable(0, 1)
	C.lua_pushcclosure(L.state, C.lua_CFunction(C.gUserdataCollector), 0)
	L.SetField(-2, "__gc")
	L.SetMetatable(-2)
}

func (L *State) PushValue(index int) {
	C.lua_pushvalue(L.state, C.int(index))
}

func (L *State) Copy(from, to int) {
	C.lua_copy(L.state, C.int(from), C.int(to))
}

func (L *State) Rotate(index, n int) {
	C.lua_rotate(L.state, C.int(index), C.int(n))
}

func (L *State) CreateTable(narr, nrec int) {
	C.lua_createtable(L.state, C.int(narr), C.int(nrec))
}

func (L *State) NewTypeMetatable(name string) {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	C.luaL_newmetatable(L.state, n)
}

func (L *State) GetTypeMetatable(name string) Type {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	return Type(C.f_luaL_getmetatable(L.state, n))
}

func (L *State) GetField(index int, name string) Type {
	s := C.CString(name)
	defer C.free(unsafe.Pointer(s))
	return Type(C.lua_getfield(L.state, C.int(index), s))
}

func (L *State) SetField(index int, name string) {
	s := C.CString(name)
	defer C.free(unsafe.Pointer(s))
	C.lua_setfield(L.state, C.int(index), s)
}

func (L *State) GetI(index, pos int) Type {
	return Type(C.lua_geti(L.state, C.int(index), C.lua_Integer(pos)))
}

func (L *State) SetI(index, pos int) {
	C.lua_seti(L.state, C.int(index), C.lua_Integer(pos))
}

func (L *State) Len(index int) int64 {
	C.lua_len(L.state, C.int(index))
	l := L.ToInteger(-1)
	L.Pop(1)
	return l
}

func (L *State) Next(index int) bool {
	return C.lua_next(L.state, C.int(index)) != 0
}

var (
	gfuncs = newStore[GFunction]()
)

//export gFunctionWrapper
func gFunctionWrapper(state *C.lua_State) C.int {
	top := C.lua_gettop(state)

	L := &State{state}

	L.GetField(int(C.f_lua_upvalueindex(1)), "id")
	id := int(C.lua_tointegerx(state, -1, nil))
	f := gfuncs.Get(id)

	C.lua_settop(state, top)

	return C.int(f(L))
}

//export gFunctionCollector
func gFunctionCollector(state *C.lua_State) C.int {
	L := &State{state}

	L.GetField(1, "id")
	id := int(C.lua_tointegerx(state, -1, nil))
	gfuncs.Pop(id)

	return 0
}

func (L *State) PushFunction(f GFunction) {
	id := gfuncs.Push(f)
	L.CreateTable(0, 1)

	C.lua_pushinteger(L.state, C.lua_Integer(id))
	L.SetField(-2, "id")

	L.CreateTable(0, 1)
	C.lua_pushcclosure(L.state, C.lua_CFunction(C.gFunctionCollector), 0)
	L.SetField(-2, "__gc")
	L.SetMetatable(-2)

	C.lua_pushcclosure(L.state, C.lua_CFunction(C.gFunctionWrapper), 1)
}

func (L *State) GetMetatable(index int) bool {
	return C.lua_getmetatable(L.state, C.int(index)) == 1
}

func (L *State) SetMetatable(index int) {
	C.lua_setmetatable(L.state, C.int(index))
}

func (L *State) GetGlobal(name string) Type {
	s := C.CString(name)
	defer C.free(unsafe.Pointer(s))
	return Type(C.lua_getglobal(L.state, s))
}

func (L *State) SetGlobal(name string) {
	s := C.CString(name)
	defer C.free(unsafe.Pointer(s))
	C.lua_setglobal(L.state, s)
}

var (
	hooks = newSyncMap[*C.lua_State, HookFunc]()
)

//export gHookWrapper
func gHookWrapper(state *C.lua_State, debug *C.lua_Debug) {
	if h, ok := hooks.Get(state); ok {
		h(&State{state}, Debug{
			Event: HookEvent(debug.event),
			l:     state,
			ar:    debug,
		})
	}
}

// SetHook sets HookFunc to the lua state.
// This method can set only one hook function per a State.
func (L *State) SetHook(mask HookMask, count int, hook HookFunc) {
	if mask == 0 || hook == nil {
		L.UnsetHook()
		return
	}

	hooks.Set(L.state, hook)
	C.lua_sethook(L.state, C.lua_Hook(C.gHookWrapper), C.int(mask), C.int(count))
}

func (L *State) UnsetHook() {
	hooks.Pop(L.state)
	C.lua_sethook(L.state, C.lua_Hook(C.gHookWrapper), 0, 0)
}

func (L *State) GetStack(level int) (d Debug, ok bool) {
	var ar C.lua_Debug
	d = Debug{
		l:  L.state,
		ar: &ar,
	}

	return d, C.lua_getstack(L.state, C.int(level), &ar) != 0
}

type LuaError struct {
	Kind    error
	Message string
}

func (e LuaError) Unwrap() error {
	return e.Kind
}

func (e LuaError) Error() string {
	return e.Message
}

func (L *State) getError(errn C.int) error {
	switch errn {
	case C.LUA_OK:
		return nil
	case C.LUA_ERRRUN:
		return LuaError{Kind: ErrRuntime, Message: L.ToString(-1)}
	case C.LUA_ERRMEM:
		return LuaError{Kind: ErrMemory, Message: L.ToString(-1)}
	case C.LUA_ERRERR:
		return LuaError{Kind: ErrError, Message: L.ToString(-1)}
	case C.LUA_ERRSYNTAX:
		return LuaError{Kind: ErrSyntax, Message: L.ToString(-1)}
	case C.LUA_YIELD:
		return ErrYield
	case C.LUA_ERRFILE:
		return LuaError{Kind: ErrFile, Message: L.ToString(-1)}
	default:
		return LuaError{Kind: ErrUnknown, Message: L.ToString(-1)}
	}
}

func (L *State) Call(nargs, nret int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok && e != nil {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	err = L.getError(C.lua_pcallk(L.state, C.int(nargs), C.int(nret), 0, 0, nil))
	return
}

var (
	readers      = newStore[io.Reader]()
	readerChunks = newSyncMap[int, unsafe.Pointer]()
)

//export gReader
func gReader(state *C.lua_State, data *C.void, size *C.size_t) *C.char {
	id := *(*int)(unsafe.Pointer(data))
	r := readers.Get(id)

	if r == nil {
		*size = 0
		return nil
	}

	if ptr, ok := readerChunks.Pop(id); ok {
		C.free(ptr)
	}

	var buf [1024 * 1024]byte
	n, err := r.Read(buf[:])
	if err == io.EOF {
		*size = 0
		return nil
	} else if err != nil {
		panic(err)
	}

	*size = C.size_t(n)

	ptr := C.CBytes(buf[:n])
	readerChunks.Set(id, ptr)
	return (*C.char)(ptr)
}

func (L *State) Load(r io.Reader, name string, isFile bool) (err error) {
	id := readers.Push(r)

	defer func() {
		readers.Pop(id)
		if r := recover(); r != nil {
			if e, ok := r.(error); ok && e != nil {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	if isFile {
		name = "@" + name
	} else {
		name = "=" + name
	}
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))

	err = L.getError(C.lua_load(L.state, C.lua_Reader(C.gReader), unsafe.Pointer(&id), n, nil))
	return
}

// WrapError wraps error by Error and set traceback.
func (L *State) WrapError(level int, err error) error {
	e := Error{
		Err: err,
	}
	if level > 0 {
		e.ChunkName, e.CurrentLine = L.Where(level)
		e.Traceback = L.Traceback(level)
	} else {
		e.Traceback = L.Traceback(1)
	}
	return e
}

// Error raises an error.
// This method throws panic of Go to catch by Call method.
func (L *State) Error(level int, err error) {
	if _, ok := err.(Error); ok {
		panic(err)
	} else {
		panic(L.WrapError(level, err))
	}
}

func (L *State) Traceback(level int) string {
	C.luaL_traceback(L.state, L.state, nil, C.int(level))
	s := L.ToString(-1)
	L.Pop(1)
	return s
}

func (L *State) Where(level int) (chunkname string, currentline int) {
	C.luaL_where(L.state, C.int(level))

	s := L.ToString(-1)
	L.Pop(1)

	if !strings.HasSuffix(s, ": ") {
		return s, 0
	}

	s2 := s[:len(s)-2]
	i := strings.LastIndex(s2, ":")
	if i < 0 {
		return s, 0
	}

	n, err := strconv.Atoi(s2[i+1:])
	if err != nil {
		return s, 0
	}

	return s[:i], n
}
