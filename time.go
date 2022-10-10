package main

import (
	"time"

	"github.com/yuin/gopher-lua"
)

func RegisterTime(L *lua.LState) {
	tbl := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"now": func(L *lua.LState) int {
			L.Push(lua.LNumber(time.Now().UnixMilli()))
			return 1
		},
		"sleep": func(L *lua.LState) int {
			time.Sleep(time.Duration(float64(L.CheckNumber(1)) * float64(time.Millisecond)))
			return 0
		},
		"format": func(L *lua.LState) int {
			L.Push(lua.LString(time.UnixMilli(int64(L.CheckNumber(1))).Format(time.RFC3339)))
			return 1
		},
	})
	L.SetField(tbl, "millisecond", lua.LNumber(1))
	L.SetField(tbl, "second", lua.LNumber(1000))
	L.SetField(tbl, "minute", lua.LNumber(60*1000))
	L.SetField(tbl, "hour", lua.LNumber(60*60*1000))
	L.SetGlobal("time", tbl)
}
