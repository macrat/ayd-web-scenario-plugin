package lua

/*
#include <stdlib.h>
#include <lua.h>

// The 'what' flags lua_getinfo
const char* what_n = "n";
const char* what_S = "S";
const char* what_l = "l";
const char* what_u = "u";
const char* what_t = "t";
const char* what_r = "r";
*/
import "C"

type HookFunc func(*State, Debug)

type Debug struct {
	Event HookEvent

	l  *C.lua_State
	ar *C.lua_Debug
}

func parseCString(s *C.char) string {
	if s == nil {
		return ""
	}
	return C.GoString(s)
}

type DebugName struct {
	Name string
	What string
}

func (d Debug) Name() DebugName {
	C.lua_getinfo(d.l, C.what_n, d.ar)
	return DebugName{
		Name: parseCString(d.ar.name),
		What: parseCString(d.ar.namewhat),
	}
}

type DebugSource struct {
	What            string
	Source          string
	Len             int
	LineDefined     int
	LastLineDefined int
	ShortSrc        string
}

func (d Debug) Source() DebugSource {
	C.lua_getinfo(d.l, C.what_S, d.ar)
	return DebugSource{
		What:            parseCString(d.ar.what),
		Source:          parseCString(d.ar.source),
		Len:             int(d.ar.srclen),
		LineDefined:     int(d.ar.linedefined),
		LastLineDefined: int(d.ar.lastlinedefined),
		ShortSrc:        parseCString(&d.ar.short_src[0]),
	}
}

func (d Debug) CurrentLine() int {
	C.lua_getinfo(d.l, C.what_l, d.ar)
	return int(d.ar.currentline)
}

type DebugUpvalues struct {
	NUps     int
	NParams  int
	IsVarArg bool
}

func (d Debug) Upvalues() DebugUpvalues {
	C.lua_getinfo(d.l, C.what_u, d.ar)
	return DebugUpvalues{
		NUps:     int(d.ar.nups),
		NParams:  int(d.ar.nparams),
		IsVarArg: d.ar.isvararg != 0,
	}
}

func (d Debug) IsTailCall() bool {
	C.lua_getinfo(d.l, C.what_t, d.ar)
	return d.ar.istailcall != 0
}

type DebugTransfer struct {
	FTransfer int
	NTransfer int
}

func (d Debug) Transfer() DebugTransfer {
	C.lua_getinfo(d.l, C.what_r, d.ar)
	return DebugTransfer{
		FTransfer: int(d.ar.ftransfer),
		NTransfer: int(d.ar.ntransfer),
	}
}

type HookMask int

const (
	MaskCall  = HookMask(C.LUA_MASKCALL)
	MaskRet   = HookMask(C.LUA_MASKRET)
	MaskLine  = HookMask(C.LUA_MASKLINE)
	MaskCount = HookMask(C.LUA_MASKCOUNT)
)

type HookEvent int

const (
	HookCall     = HookEvent(C.LUA_HOOKCALL)
	HookRet      = HookEvent(C.LUA_HOOKRET)
	HookTailCall = HookEvent(C.LUA_HOOKTAILCALL)
	HookLine     = HookEvent(C.LUA_HOOKLINE)
	HookCount    = HookEvent(C.LUA_HOOKCOUNT)
)

func (e HookEvent) String() string {
	switch e {
	case HookCall:
		return "call"
	case HookRet:
		return "ret"
	case HookTailCall:
		return "tailcall"
	case HookLine:
		return "line"
	case HookCount:
		return "count"
	default:
		return "unknown"
	}
}
