package main

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/yuin/gopher-lua"
)

type ElementsArray []Element

func NewElementsArray(L *lua.LState, t *Tab, query string) ElementsArray {
	var ids []cdp.NodeID
	t.RunSelector(query, chromedp.NodeIDs(query, &ids, chromedp.ByQueryAll, chromedp.AtLeast(0)))

	var es ElementsArray
	for _, id := range ids {
		e := Element{
			query: query,
			ids:   []cdp.NodeID{id},
			tab:   t,
		}
		es = append(es, e)
	}
	return es
}

func (es ElementsArray) ToLua(L *lua.LState) *lua.LTable {
	tbl := L.NewTable()
	for _, e := range es {
		tbl.Append(e.ToLua(L))
	}
	L.SetMetatable(tbl, L.GetTypeMetatable("elementsarray"))
	return tbl
}

func CheckElementsArray(L *lua.LState) ElementsArray {
	t := L.CheckTable(1)

	var es ElementsArray

	idx := lua.LNil
	for {
		var v lua.LValue
		idx, v = t.Next(idx)
		if v == lua.LNil {
			break
		} else if v, ok := v.(*lua.LUserData); !ok {
			L.ArgError(1, "array of elements expected.")
		} else if e, ok := v.Value.(Element); !ok {
			L.ArgError(1, "array of elements expected.")
		} else {
			es = append(es, e)
		}
	}

	return es
}

func RegisterElementsArrayType(ctx context.Context, L *lua.LState) {
	runAll := func(f func(Element, *lua.LState)) *lua.LFunction {
		return L.NewFunction(func(L *lua.LState) int {
			es := CheckElementsArray(L)
			for _, e := range es {
				f(e, L)
			}
			L.Push(L.Get(1))
			return 1
		})
	}

	getAll := func(f func(Element, *lua.LState) int) func(ElementsArray, *lua.LState) int {
		return func(es ElementsArray, L *lua.LState) int {
			rs := L.NewTable()
			for _, e := range es {
				n := f(e, L)
				for i := 0; i < n; i++ {
					rs.Append(L.Get(-1))
					L.Pop(1)
				}
			}
			L.Push(rs)
			return 1
		}
	}

	methods := map[string]*lua.LFunction{
		"all": L.NewFunction(func(L *lua.LState) int {
			query := L.CheckString(2)
			var es ElementsArray
			for _, e := range CheckElementsArray(L) {
				es = append(es, e.SelectAll(L, query)...)
			}
			L.Push(es.ToLua(L))
			return 1
		}),
		"sendKeys": runAll(Element.SendKeys),
		"setValue": runAll(Element.SetValue),
		"click":    runAll(Element.Click),
		"submit":   runAll(Element.Submit),
	}

	getters := map[string]func(ElementsArray, *lua.LState) int{
		"text":      getAll(Element.GetText),
		"innerHTML": getAll(Element.GetInnerHTML),
		"outerHTML": getAll(Element.GetOuterHTML),
		"value":     getAll(Element.GetValue),
	}

	query := L.SetFuncs(L.NewTypeMetatable("elementsarray"), map[string]lua.LGFunction{
		"__call": func(L *lua.LState) int {
			query := L.CheckString(2)
			for _, e := range CheckElementsArray(L) {
				es := e.SelectAll(L, query)
				if len(es) > 0 {
					L.Push(es[0].ToLua(L))
					return 1
				}
			}
			L.RaiseError("no such element")
			return 0
		},
		"__index": func(L *lua.LState) int {
			name := L.CheckString(2)

			if f, ok := getters[name]; ok {
				return f(CheckElementsArray(L), L)
			} else if f, ok := methods[name]; ok {
				L.Push(f)
				return 1
			} else {
				return 0
			}
		},
		"__tostring": func(L *lua.LState) int {
			es := CheckElementsArray(L)
			L.Push(lua.LString(fmt.Sprintf("[%d elements]", len(es))))
			return 1
		},
	})
	L.SetGlobal("elementsarray", query)
}
