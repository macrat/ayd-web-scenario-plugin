package main

import (
	"time"

	"github.com/yuin/gopher-lua"
)

func RegisterTime(L *lua.LState) {
	L.SetGlobal("time", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"sleep": func(L *lua.LState) int {
			time.Sleep(time.Duration(L.CheckInt64(1)) * time.Millisecond)
			return 0
		},
	}))
}
