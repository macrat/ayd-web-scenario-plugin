package webscenario

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/yuin/gopher-lua"
)

type Element struct {
	name string
	node *cdp.Node
	tab  *Tab
}

func nodeAction(sel interface{}, node **cdp.Node, opts ...chromedp.QueryOption) chromedp.QueryAction {
	return chromedp.QueryAfter(sel, func(ctx context.Context, id runtime.ExecutionContextID, nodes ...*cdp.Node) error {
		*node = nodes[0]
		return nil
	}, opts...)
}

func NewElement(L *lua.LState, t *Tab, query string) Element {
	var node *cdp.Node
	name := fmt.Sprintf("$(%q)", strings.TrimSpace(query))
	t.RunSelector(L, name, nodeAction(query, &node, chromedp.ByQuery))

	return Element{
		name: name,
		node: node,
		tab:  t,
	}
}

func (e Element) ToLua(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = e
	L.SetMetatable(ud, L.GetTypeMetatable("element"))
	return ud
}

func newElementsTableFromNodes(L *lua.LState, t *Tab, name string, nodes []*cdp.Node) *lua.LTable {
	tbl := L.NewTable()
	for _, node := range nodes {
		tbl.Append(Element{
			name: name,
			node: node,
			tab:  t,
		}.ToLua(L))
	}

	idx := 1
	L.SetMetatable(tbl, L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"__call": func(L *lua.LState) int {
			if idx > len(nodes) {
				L.Push(lua.LNil)
			} else {
				L.Push(tbl.RawGet(lua.LNumber(idx)))
				idx++
			}
			return 1
		},
	}))

	return tbl
}

func NewElementsTable(L *lua.LState, t *Tab, query string) *lua.LTable {
	var nodes []*cdp.Node
	name := fmt.Sprintf("$:all(%q)", strings.TrimSpace(query))
	t.RunSelector(
		L,
		name,
		chromedp.Nodes(query, &nodes, chromedp.ByQueryAll, chromedp.AtLeast(0)),
	)
	return newElementsTableFromNodes(L, t, query, nodes)
}

func NewElementsTableByXPath(L *lua.LState, t *Tab, query string) *lua.LTable {
	var nodes []*cdp.Node
	name := fmt.Sprintf("$:xpath(%q)", strings.TrimSpace(query))
	t.RunSelector(L, name, chromedp.Nodes(query, &nodes, chromedp.BySearch))
	return newElementsTableFromNodes(L, t, name, nodes)
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

func (e Element) ids() []cdp.NodeID {
	return []cdp.NodeID{e.node.NodeID}
}

func (e Element) Select(L *lua.LState, query string) Element {
	name := fmt.Sprintf("%s(%q)", e.name, strings.TrimSpace(query))

	var node *cdp.Node
	e.tab.Run(
		L,
		name,
		false,
		nodeAction(query, &node, chromedp.ByQuery, chromedp.FromNode(e.node)),
	)

	return Element{
		name: name,
		node: node,
		tab:  e.tab,
	}
}

func (e Element) SelectAll(L *lua.LState, query string) *lua.LTable {
	name := fmt.Sprintf("%s:all(%q)", e.name, strings.TrimSpace(query))

	var nodes []*cdp.Node

	e.tab.RunSelector(
		L,
		name,
		chromedp.Nodes(
			query,
			&nodes,
			chromedp.ByQueryAll,
			chromedp.FromNode(e.node),
			chromedp.AtLeast(0),
		),
	)

	return newElementsTableFromNodes(L, e.tab, name, nodes)
}

func (e Element) SendKeys(L *lua.LState) {
	text := L.CheckString(2)
	e.tab.Run(L, fmt.Sprintf("%s:sendKeys(%q)", e.name, text), true, chromedp.SendKeys(e.ids(), text, chromedp.ByNodeID))
}

func (e Element) SetValue(L *lua.LState) {
	value := L.CheckString(2)
	e.tab.Run(L, fmt.Sprintf("%s:setValue(%q)", e.name, value), true, chromedp.SetValue(e.ids(), value, chromedp.ByNodeID))
}

func (e Element) Click(L *lua.LState) {
	e.tab.Run(L, fmt.Sprintf("%s:click()", e.name), true, chromedp.Click(e.ids(), chromedp.ByNodeID))
}

func (e Element) Submit(L *lua.LState) {
	e.tab.Run(L, fmt.Sprintf("%s:submit()", e.name), true, chromedp.Submit(e.ids(), chromedp.ByNodeID))
}

func (e Element) Focus(L *lua.LState) {
	e.tab.Run(L, fmt.Sprintf("%s:focus()", e.name), false, chromedp.Focus(e.ids(), chromedp.ByNodeID))
}

func (e Element) Blur(L *lua.LState) {
	e.tab.Run(L, fmt.Sprintf("%s:blur()", e.name), false, chromedp.Blur(e.ids(), chromedp.ByNodeID))
}

func (e Element) Screenshot(L *lua.LState) {
	name := L.ToString(2)

	var buf []byte
	e.tab.Run(
		L,
		fmt.Sprintf("%s:screenshot(%v)", e.name, name),
		false,
		chromedp.Screenshot(e.ids(), &buf, chromedp.ByNodeID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return e.tab.Save(name, ".jpg", buf)
		}),
	)
}

func (e Element) GetText(L *lua.LState) int {
	var text string
	e.tab.Run(L, fmt.Sprintf("%s.text", e.name), false, chromedp.Text(e.ids(), &text, chromedp.ByNodeID))
	L.Push(lua.LString(text))
	return 1
}

func (e Element) GetInnerHTML(L *lua.LState) int {
	var html string
	e.tab.Run(L, fmt.Sprintf("%s.innerHTML", e.name), false, chromedp.InnerHTML(e.ids(), &html, chromedp.ByNodeID))
	L.Push(lua.LString(html))
	return 1
}

func (e Element) GetOuterHTML(L *lua.LState) int {
	var html string
	e.tab.Run(L, fmt.Sprintf("%s.outerHTML", e.name), false, chromedp.OuterHTML(e.ids(), &html, chromedp.ByNodeID))
	L.Push(lua.LString(html))
	return 1
}

func (e Element) GetValue(L *lua.LState) int {
	var value string
	e.tab.Run(L, fmt.Sprintf("%s.value", e.name), false, chromedp.Value(e.ids(), &value, chromedp.ByNodeID))
	L.Push(lua.LString(value))
	return 1
}

func (e Element) GetAttribute(L *lua.LState) int {
	name := L.CheckString(2)

	var value string
	var ok bool
	e.tab.Run(L, fmt.Sprintf("%s[%q]", e.name, name), false, chromedp.AttributeValue(e.ids(), name, &value, &ok, chromedp.ByNodeID))

	if ok {
		L.Push(lua.LString(value))
		return 1
	} else {
		return 0
	}
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
		"all": L.NewFunction(func(L *lua.LState) int {
			e := CheckElement(L)
			query := L.CheckString(2)
			L.Push(e.SelectAll(L, query))
			return 1
		}),
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
		"__call": func(L *lua.LState) int {
			e := CheckElement(L)
			query := L.CheckString(2)
			L.Push(e.Select(L, query).ToLua(L))
			return 1
		},
		"__index": func(L *lua.LState) int {
			name := L.CheckString(2)

			if f, ok := getters[name]; ok {
				return f(CheckElement(L), L)
			} else if f, ok := methods[name]; ok {
				L.Push(f)
				return 1
			} else {
				return CheckElement(L).GetAttribute(L)
			}
		},
		"__tostring": func(L *lua.LState) int {
			e := CheckElement(L)
			L.Push(lua.LString(fmt.Sprintf("%s", e.name)))
			return 1
		},
	})
	L.SetGlobal("element", query)
}
