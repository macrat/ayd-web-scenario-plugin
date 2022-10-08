package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/yuin/gopher-lua"
)

type Element struct {
	query string
	ids   []cdp.NodeID
	tab   *Tab
}

func NewElement(L *lua.LState, t *Tab, query string) *Element {
	var ids []cdp.NodeID
	t.Run(L, chromedp.NodeIDs(query, &ids, chromedp.ByQuery))

	return &Element{
		query: query,
		ids:   ids,
		tab:   t,
	}
}

func (e *Element) ToLua(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = e
	L.SetMetatable(ud, L.GetTypeMetatable("element"))
	return ud
}

func NewElementArray(L *lua.LState, t *Tab, query string) *lua.LTable {
	var ids []cdp.NodeID
	t.Run(L, chromedp.NodeIDs(query, &ids, chromedp.ByQueryAll))

	es := L.NewTable()
	for _, id := range ids {
		e := Element{
			query: query,
			ids:   []cdp.NodeID{id},
			tab:   t,
		}
		es.Append(e.ToLua(L))
	}
	return es
}

func CheckElement(L *lua.LState) *Element {
	if e, ok := L.CheckUserData(1).Value.(*Element); ok {
		return e
	}
	L.ArgError(1, "element expected. perhaps you call it like tab().xxx() instead of tab():xxx().")
	return nil
}

func RegisterElementType(ctx context.Context, L *lua.LState) {
	methods := map[string]*lua.LFunction{
		"wait": L.NewFunction(func(L *lua.LState) int {
			e := CheckElement(L)
			fmt.Printf("waiting for %s\n", e.query)
			CheckTab(L).Run(L, chromedp.WaitReady(e.ids, chromedp.ByNodeID))
			return 0
		}),
		"sendKeys": L.NewFunction(func(L *lua.LState) int {
			e := CheckElement(L)
			text := L.ToString(2)
			fmt.Printf("send keys %s\n", e.query)
			e.tab.Run(L, chromedp.SendKeys(e.ids, text, chromedp.ByNodeID))

			L.Push(L.Get(1))
			return 1
		}),
		"click": L.NewFunction(func(L *lua.LState) int {
			e := CheckElement(L)
			fmt.Printf("click on %s\n", e.query)
			e.tab.Run(L, chromedp.Click(e.ids, chromedp.ByNodeID))

			L.Push(L.Get(1))
			return 1
		}),
		"setValue": L.NewFunction(func(L *lua.LState) int {
			e := CheckElement(L)
			v := L.ToString(3)
			fmt.Printf("set value %q to %s\n", v, e.query)
			e.tab.Run(L, chromedp.SetValue(e.ids, v, chromedp.ByNodeID))

			L.Push(L.Get(1))
			return 1
		}),
		"submit": L.NewFunction(func(L *lua.LState) int {
			e := CheckElement(L)
			fmt.Printf("submit %s\n", e.query)
			e.tab.Run(L, chromedp.Submit(e.ids, chromedp.ByNodeID))

			L.Push(L.Get(1))
			return 1
		}),
		"focus": L.NewFunction(func(L *lua.LState) int {
			e := CheckElement(L)
			fmt.Printf("focus on %s\n", e.query)
			e.tab.Run(L, chromedp.Focus(e.ids, chromedp.ByNodeID))

			L.Push(L.Get(1))
			return 1
		}),
		"blur": L.NewFunction(func(L *lua.LState) int {
			e := CheckElement(L)
			fmt.Printf("blur from %s\n", e.query)
			e.tab.Run(L, chromedp.Blur(e.ids, chromedp.ByNodeID))

			L.Push(L.Get(1))
			return 1
		}),
		"screenshot": L.NewFunction(func(L *lua.LState) int {
			var buf []byte
			e := CheckElement(L)
			name := L.CheckString(2)
			e.tab.Run(L, chromedp.Screenshot(e.ids, &buf, chromedp.ByNodeID))
			os.WriteFile(name, buf, 0644)

			L.Push(L.Get(1))
			return 1
		}),
	}

	query := L.SetFuncs(L.NewTypeMetatable("element"), map[string]lua.LGFunction{
		"__index": func(L *lua.LState) int {
			name := L.CheckString(2)

			switch name {
			case "text":
				var text string
				e := CheckElement(L)
				fmt.Printf("get inner text of %s\n", e.query)
				e.tab.Run(L, chromedp.Text(e.ids, &text, chromedp.ByNodeID))
				L.Push(lua.LString(text))
				return 1
			case "innerHTML":
				var html string
				e := CheckElement(L)
				fmt.Printf("get inner html of %s\n", e.query)
				e.tab.Run(L, chromedp.InnerHTML(e.ids, &html, chromedp.ByNodeID))
				L.Push(lua.LString(html))
				return 1
			case "outerHTML":
				var html string
				e := CheckElement(L)
				fmt.Printf("get outer html of %s\n", e.query)
				e.tab.Run(L, chromedp.OuterHTML(e.ids, &html, chromedp.ByNodeID))
				L.Push(lua.LString(html))
				return 1
			case "value":
				var value string
				e := CheckElement(L)
				fmt.Printf("get value of %s\n", e.query)
				e.tab.Run(L, chromedp.Value(e.ids, &value, chromedp.ByNodeID))
				L.Push(lua.LString(value))
				return 1
			default:
				if f, ok := methods[name]; ok {
					L.Push(f)
					return 1
				} else {
					return 0
				}
			}
		},
		"__tostring": func(L *lua.LState) int {
			e := CheckElement(L)
			L.Push(lua.LString(fmt.Sprintf("[%d elements]{%s}", len(e.ids), e.query)))
			return 1
		},
	})
	L.SetGlobal("element", query)
}
