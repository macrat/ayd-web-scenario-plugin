package webscenario

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func RegisterAssert(L *lua.LState) {
	failed := func(operator string, a, b string) {
		L.RaiseError("assertion failed: %s %s %s", a, operator, b)
	}

	order := func(operator string, s func(a, b string) bool, n func(a, b float64) bool) lua.LGFunction {
		return func(L *lua.LState) int {
			a, b := L.Get(1), L.Get(2)
			if a.Type() != b.Type() {
				failed(operator, LValueToString(a), LValueToString(b))
			}
			switch a.Type() {
			case lua.LTString:
				if !s(string(a.(lua.LString)), string(b.(lua.LString))) {
					failed(operator, LValueToString(a), LValueToString(b))
				}
			case lua.LTNumber:
				if !n(float64(a.(lua.LNumber)), float64(b.(lua.LNumber))) {
					failed(operator, LValueToString(a), LValueToString(b))
				}
			default:
				failed(operator, LValueToString(a), LValueToString(b))
			}
			return 2
		}
	}

	tbl := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"eq": func(L *lua.LState) int {
			a := UnpackLValue(L.Get(1))
			b := UnpackLValue(L.Get(2))
			if !reflect.DeepEqual(a, b) {
				failed("==", LValueToString(L.Get(1)), LValueToString(L.Get(2)))
			}
			return 2
		},
		"ne": func(L *lua.LState) int {
			a := UnpackLValue(L.Get(1))
			b := UnpackLValue(L.Get(2))
			if reflect.DeepEqual(a, b) {
				failed("~=", LValueToString(L.Get(1)), LValueToString(L.Get(2)))
			}
			return 2
		},
		"lt": order("<", func(a, b string) bool { return a < b }, func(a, b float64) bool { return a < b }),
		"le": order("<=", func(a, b string) bool { return a <= b }, func(a, b float64) bool { return a <= b }),
		"gt": order(">", func(a, b string) bool { return a > b }, func(a, b float64) bool { return a > b }),
		"ge": order(">=", func(a, b string) bool { return a >= b }, func(a, b float64) bool { return a >= b }),
	})

	meta := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"__call": func(L *lua.LState) int {
			if !L.ToBool(2) {
				L.RaiseError("%s", L.OptString(3, "assertion failed!"))
			}
			return L.GetTop() - 1
		},
		"__tostring": func(L *lua.LState) int {
			L.Push(lua.LString("assert"))
			return 1
		},
	})
	L.SetMetatable(tbl, meta)

	L.SetGlobal("assert", tbl)
}
