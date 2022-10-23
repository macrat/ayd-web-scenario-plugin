package main

import (
	"time"

	"github.com/yuin/gopher-lua"
)

func RegisterTime(env *Environment) {
	env.RegisterNewType("time", map[string]lua.LGFunction{
		"now": func(L *lua.LState) int {
			env.Yield()
			L.Push(lua.LNumber(time.Now().UnixMilli()))
			return 1
		},
		"sleep": func(L *lua.LState) int {
			AsyncRun(env, func() struct{} {
				time.Sleep(time.Duration(float64(L.CheckNumber(1)) * float64(time.Millisecond)))
				return struct{}{}
			})
			return 0
		},
		"format": func(L *lua.LState) int {
			env.Yield()
			L.Push(lua.LString(time.UnixMilli(int64(L.CheckNumber(1))).Format(time.RFC3339)))
			return 1
		},
	}, map[string]lua.LValue{
		"millisecond": lua.LNumber(1),
		"second":      lua.LNumber(1000),
		"minute":      lua.LNumber(60 * 1000),
		"hour":        lua.LNumber(60 * 60 * 1000),
	})
}
