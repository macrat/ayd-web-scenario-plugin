package webscenario

import (
	"encoding/json"

	"github.com/yuin/gopher-lua"
)

type Encodings Environment

func (e *Encodings) ToJSON(L *lua.LState) int {
	bs, err := json.Marshal(UnpackLValue(L.Get(1)))
	(*Environment)(e).HandleError(err)
	L.Push(lua.LString(string(bs)))
	return 1
}

func (e *Encodings) FromJSON(L *lua.LState) int {
	var v any
	json.Unmarshal([]byte(L.CheckString(1)), &v)
	L.Push(PackLValue(L, v))
	return 1
}

func RegisterEncodings(env *Environment) {
	env.RegisterFunction("tojson", (*Encodings)(env).ToJSON)
	env.RegisterFunction("fromjson", (*Encodings)(env).FromJSON)
}
