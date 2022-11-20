package webscenario

import (
	"encoding/json"
	"reflect"

	"github.com/macrat/ayd-web-scenario/internal/lua"
)

func RegisterAssert(L *lua.State) {
	failed := func(L *lua.State, operator string) {
		a, err := json.Marshal(L.ToAny(1))
		if err != nil {
			a = []byte(L.ToString(1))
		}

		b, err := json.Marshal(L.ToAny(2))
		if err != nil {
			b = []byte(L.ToString(2))
		}

		L.Errorf(1, "assertion failed: %s %s %s", a, operator, b)
	}

	order := func(operator string, s func(a, b string) bool, n func(a, b float64) bool) lua.GFunction {
		return func(L *lua.State) int {
			if L.Type(1) != L.Type(2) {
				failed(L, operator)
			}
			switch L.Type(1) {
			case lua.String:
				a, b := L.ToString(1), L.ToString(2)
				if !s(a, b) {
					failed(L, operator)
				}
			case lua.Number:
				if !n(L.ToNumber(1), L.ToNumber(2)) {
					failed(L, operator)
				}
			default:
				failed(L, operator)
			}
			return 2
		}
	}

	L.CreateTable(0, 6)
	L.SetFuncs(-1, map[string]lua.GFunction{
		"eq": func(L *lua.State) int {
			a := L.ToAny(1)
			b := L.ToAny(2)
			if !reflect.DeepEqual(a, b) {
				failed(L, "==")
			}
			return 2
		},
		"ne": func(L *lua.State) int {
			a := L.ToAny(1)
			b := L.ToAny(2)
			if reflect.DeepEqual(a, b) {
				failed(L, "~=")
			}
			return 2
		},
		"lt": order("<", func(a, b string) bool { return a < b }, func(a, b float64) bool { return a < b }),
		"le": order("<=", func(a, b string) bool { return a <= b }, func(a, b float64) bool { return a <= b }),
		"gt": order(">", func(a, b string) bool { return a > b }, func(a, b float64) bool { return a > b }),
		"ge": order(">=", func(a, b string) bool { return a >= b }, func(a, b float64) bool { return a >= b }),
	})

	L.CreateTable(0, 2)
	L.SetFunction(-1, "__call", func(L *lua.State) int {
		if !L.ToBoolean(2) {
			msg := "assertion failed!"
			if L.Type(3) != lua.Nil {
				msg = L.ToString(3)
			}
			L.Errorf(1, "%s", msg)
		}
		return L.GetTop() - 1
	})
	L.SetMetatable(-2)

	L.SetGlobal("assert")
}
