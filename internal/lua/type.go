package lua

/*
#include <lua.h>
*/
import "C"

type GFunction func(*State) int

type Type int

const (
	Boolean       = Type(C.LUA_TBOOLEAN)
	Function      = Type(C.LUA_TFUNCTION)
	LightUserdata = Type(C.LUA_TLIGHTUSERDATA)
	Nil           = Type(C.LUA_TNIL)
	None          = Type(C.LUA_TNONE)
	Number        = Type(C.LUA_TNUMBER)
	String        = Type(C.LUA_TSTRING)
	Table         = Type(C.LUA_TTABLE)
	Thread        = Type(C.LUA_TTHREAD)
	Userdata      = Type(C.LUA_TUSERDATA)
)

func (t Type) String() string {
	switch t {
	case Boolean:
		return "boolean"
	case Function:
		return "function"
	case LightUserdata:
		return "lightuserdata"
	case Nil:
		return "nil"
	case None:
		return "none"
	case Number:
		return "number"
	case String:
		return "string"
	case Table:
		return "table"
	case Thread:
		return "thread"
	case Userdata:
		return "userdata"
	default:
		return "unknown"
	}
}
