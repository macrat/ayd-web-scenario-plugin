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

func NewElement(L *lua.LState, t *Tab, query string) Element {
	var ids []cdp.NodeID
	t.Run(L, chromedp.NodeIDs(query, &ids, chromedp.ByQuery))

	return Element{
		query: query,
		ids:   ids,
		tab:   t,
	}
}

func (e Element) ToLua(L *lua.LState) *lua.LUserData {
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

func CheckElement(L *lua.LState) Element {
	if ud, ok := L.Get(1).(*lua.LUserData); ok {
		if e, ok := ud.Value.(Element); ok {
			return e
		}
	}

	L.ArgError(1, "element expected. perhaps you call it like tab().xxx() instead of tab():xxx().")
	return Element{}
}

func (e Element) Wait(L *lua.LState) {
	fmt.Printf("waiting for %s\n", e.query)
	e.tab.Run(L, chromedp.WaitReady(e.ids, chromedp.ByNodeID))
}

func (e Element) SendKeys(L *lua.LState) {
	text := L.ToString(2)

	fmt.Printf("send keys %s\n", e.query)
	e.tab.Run(L, chromedp.SendKeys(e.ids, text, chromedp.ByNodeID))
}

func (e Element) SetValue(L *lua.LState) {
	value := L.ToString(3)

	fmt.Printf("set value %q to %s\n", value, e.query)
	e.tab.Run(L, chromedp.SetValue(e.ids, value, chromedp.ByNodeID))
}

func (e Element) Click(L *lua.LState) {
	fmt.Printf("click on %s\n", e.query)
	e.tab.Run(L, chromedp.Click(e.ids, chromedp.ByNodeID))
}

func (e Element) Submit(L *lua.LState) {
	fmt.Printf("submit %s\n", e.query)
	e.tab.Run(L, chromedp.Submit(e.ids, chromedp.ByNodeID))
}

func (e Element) Focus(L *lua.LState) {
	fmt.Printf("focus on %s\n", e.query)
	e.tab.Run(L, chromedp.Focus(e.ids, chromedp.ByNodeID))
}

func (e Element) Blur(L *lua.LState) {
	fmt.Printf("blur from %s\n", e.query)
	e.tab.Run(L, chromedp.Blur(e.ids, chromedp.ByNodeID))
}

func (e Element) Screenshot(L *lua.LState) {
	name := L.CheckString(2)

	var buf []byte
	fmt.Printf("take a screenshot of %s\n", e.query)
	e.tab.Run(L, chromedp.Screenshot(e.ids, &buf, chromedp.ByNodeID))
	os.WriteFile(name, buf, 0644)
}

func (e Element) GetText(L *lua.LState) int {
	var text string
	fmt.Printf("get inner text of %s\n", e.query)
	e.tab.Run(L, chromedp.Text(e.ids, &text, chromedp.ByNodeID))
	L.Push(lua.LString(text))
	return 1
}

func (e Element) GetInnerHTML(L *lua.LState) int {
	var html string
	fmt.Printf("get inner html of %s\n", e.query)
	e.tab.Run(L, chromedp.InnerHTML(e.ids, &html, chromedp.ByNodeID))
	L.Push(lua.LString(html))
	return 1
}

func (e Element) GetOuterHTML(L *lua.LState) int {
	var html string
	fmt.Printf("get outer html of %s\n", e.query)
	e.tab.Run(L, chromedp.OuterHTML(e.ids, &html, chromedp.ByNodeID))
	L.Push(lua.LString(html))
	return 1
}

func (e Element) GetValue(L *lua.LState) int {
	var value string
	fmt.Printf("get value of %s\n", e.query)
	e.tab.Run(L, chromedp.Value(e.ids, &value, chromedp.ByNodeID))
	L.Push(lua.LString(value))
	return 1
}

func RegisterElementType(ctx context.Context, L *lua.LState) {
	fn := func(f func(Element, *lua.LState)) *lua.LFunction {
		return L.NewFunction(func(L *lua.LState) int {
			f(CheckElement(L), L)
			L.Push(L.Get(1))
			return 1
		})
	}

	methods := map[string]*lua.LFunction{
		"wait":       fn(Element.Wait),
		"sendKeys":   fn(Element.SendKeys),
		"setValue":   fn(Element.SetValue),
		"click":      fn(Element.Click),
		"submit":     fn(Element.Submit),
		"focus":      fn(Element.Focus),
		"blur":       fn(Element.Blur),
		"screenshot": fn(Element.Screenshot),
	}

	getters := map[string]func(Element, *lua.LState) int{
		"text":      Element.GetText,
		"innerHTML": Element.GetInnerHTML,
		"outerHTML": Element.GetOuterHTML,
		"value":     Element.GetValue,
	}

	query := L.SetFuncs(L.NewTypeMetatable("element"), map[string]lua.LGFunction{
		"__index": func(L *lua.LState) int {
			name := L.CheckString(2)

			if f, ok := getters[name]; ok {
				return f(CheckElement(L), L)
			} else if f, ok := methods[name]; ok {
				L.Push(f)
				return 1
			} else {
				return 0
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
