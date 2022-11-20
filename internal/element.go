package webscenario

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/macrat/ayd-web-scenario/internal/lua"
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

func NewElement(L *lua.State, t *Tab, query string) Element {
	var node *cdp.Node
	name := fmt.Sprintf("$(%q)", strings.TrimSpace(query))
	t.RunSelector(L, name, nodeAction(query, &node, chromedp.ByQuery))

	return Element{
		name: name,
		node: node,
		tab:  t,
	}
}

func (e Element) PushTo(L *lua.State) {
	L.PushUserdata(e)
	L.GetTypeMetatable("element")
	L.SetMetatable(-2)
	return
}

func pushElementsTableFromNodes(L *lua.State, t *Tab, name string, nodes []*cdp.Node) {
	L.CreateTable(len(nodes), 0)

	for i, node := range nodes {
		Element{
			name: name,
			node: node,
			tab:  t,
		}.PushTo(L)
		L.SetI(-2, i+1)
	}

	idx := 1
	L.CreateTable(0, 1)
	L.SetFuncs(-1, map[string]lua.GFunction{
		"__call": func(L *lua.State) int {
			if idx > len(nodes) {
				L.PushNil()
			} else {
				L.GetI(1, idx)
				idx++
			}
			return 1
		},
	})
	L.SetMetatable(-2)
}

func PushElementsTable(L *lua.State, t *Tab, query string) {
	var nodes []*cdp.Node
	name := fmt.Sprintf("$:all(%q)", strings.TrimSpace(query))
	t.RunSelector(
		L,
		name,
		chromedp.Nodes(query, &nodes, chromedp.ByQueryAll, chromedp.AtLeast(0)),
	)
	pushElementsTableFromNodes(L, t, query, nodes)
}

func PushElementsTableByXPath(L *lua.State, t *Tab, query string) {
	var nodes []*cdp.Node
	name := fmt.Sprintf("$:xpath(%q)", strings.TrimSpace(query))
	t.RunSelector(L, name, chromedp.Nodes(query, &nodes, chromedp.BySearch))
	pushElementsTableFromNodes(L, t, name, nodes)
}

func CheckElement(L *lua.State) Element {
	if e, ok := L.ToUserdata(1).(Element); ok {
		return e
	}

	L.ArgErrorf(1, "element expected, but got %s. perhaps you call it like tab().xxx() instead of tab():xxx().", L.Type(1))
	return Element{}
}

func (e Element) ids() []cdp.NodeID {
	return []cdp.NodeID{e.node.NodeID}
}

func (e Element) Select(L *lua.State, query string) {
	name := fmt.Sprintf("%s(%q)", e.name, strings.TrimSpace(query))

	var node *cdp.Node
	e.tab.Run(
		L,
		name,
		false,
		0,
		nodeAction(query, &node, chromedp.ByQuery, chromedp.FromNode(e.node)),
	)

	Element{
		name: name,
		node: node,
		tab:  e.tab,
	}.PushTo(L)
}

func (e Element) SelectAll(L *lua.State, query string) {
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

	pushElementsTableFromNodes(L, e.tab, name, nodes)
}

func (e Element) SendKeys(L *lua.State) {
	text := L.CheckString(2)

	mod := input.ModifierNone

	if L.Type(3) == lua.Table {
		L.PushNil()

		for L.Next(3) {
			switch L.ToString(-1) {
			case "alt":
				mod |= input.ModifierAlt
			case "ctrl":
				mod |= input.ModifierCtrl
			case "meta":
				mod |= input.ModifierMeta
			case "shift":
				mod |= input.ModifierShift
			}
			L.Pop(1)
		}
	}

	e.tab.Run(
		L,
		fmt.Sprintf("%s:sendKeys(%q)", e.name, text),
		true,
		0,
		chromedp.KeyEventNode(e.node, text, chromedp.KeyModifiers(mod)),
	)
}

func (e Element) SetValue(L *lua.State) {
	value := L.CheckString(2)
	e.tab.Run(L, fmt.Sprintf("%s:setValue(%q)", e.name, value), true, 0, chromedp.SetValue(e.ids(), value, chromedp.ByNodeID))
}

func (e Element) Click(L *lua.State) {
	var button string

	if typ := L.Type(2); typ == lua.String {
		button = L.ToString(2)
	} else if typ == lua.Nil {
		button = "left"
	} else {
		L.ArgErrorf(2, "string or nil expected, got %s", typ)
	}

	var name string
	if button == "left" {
		name = fmt.Sprintf("%s:click()", e.name)
	} else {
		name = fmt.Sprintf("%s:click(%q)", e.name, button)
	}

	e.tab.Run(L, name, true, 0, chromedp.MouseClickNode(e.node, chromedp.Button(button)))
}

func (e Element) Submit(L *lua.State) {
	e.tab.Run(L, fmt.Sprintf("%s:submit()", e.name), true, 0, chromedp.Submit(e.ids(), chromedp.ByNodeID))
}

func (e Element) Focus(L *lua.State) {
	e.tab.Run(L, fmt.Sprintf("%s:focus()", e.name), false, 0, chromedp.Focus(e.ids(), chromedp.ByNodeID))
}

func (e Element) Blur(L *lua.State) {
	e.tab.Run(L, fmt.Sprintf("%s:blur()", e.name), false, 0, chromedp.Blur(e.ids(), chromedp.ByNodeID))
}

func (e Element) Screenshot(L *lua.State) {
	name := L.ToString(2)

	var buf []byte
	e.tab.Run(
		L,
		fmt.Sprintf("%s:screenshot(%v)", e.name, name),
		false,
		0,
		chromedp.Screenshot(e.ids(), &buf, chromedp.ByNodeID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return e.tab.Save(name, ".png", buf)
		}),
	)
}

func (e Element) GetText(L *lua.State) int {
	var text string
	e.tab.Run(L, fmt.Sprintf("%s.text", e.name), false, 0, chromedp.Text(e.ids(), &text, chromedp.ByNodeID))
	L.PushString(text)
	return 1
}

func (e Element) GetInnerHTML(L *lua.State) int {
	var html string
	e.tab.Run(L, fmt.Sprintf("%s.innerHTML", e.name), false, 0, chromedp.InnerHTML(e.ids(), &html, chromedp.ByNodeID))
	L.PushString(html)
	return 1
}

func (e Element) GetOuterHTML(L *lua.State) int {
	var html string
	e.tab.Run(L, fmt.Sprintf("%s.outerHTML", e.name), false, 0, chromedp.OuterHTML(e.ids(), &html, chromedp.ByNodeID))
	L.PushString(html)
	return 1
}

func (e Element) GetValue(L *lua.State) int {
	var value string
	e.tab.Run(L, fmt.Sprintf("%s.value", e.name), false, 0, chromedp.Value(e.ids(), &value, chromedp.ByNodeID))
	L.PushString(value)
	return 1
}

func (e Element) GetAttribute(L *lua.State) int {
	name := L.CheckString(2)

	var value string
	var ok bool
	e.tab.Run(L, fmt.Sprintf("%s[%q]", e.name, name), false, 0, chromedp.AttributeValue(e.ids(), name, &value, &ok, chromedp.ByNodeID))

	if ok {
		L.PushString(value)
		return 1
	} else {
		return 0
	}
}

func RegisterElementType(ctx context.Context, L *lua.State) {
	fn := func(f func(Element, *lua.State)) lua.GFunction {
		return func(L *lua.State) int {
			f(CheckElement(L), L)
			L.SetTop(1)
			return 1
		}
	}

	L.NewTypeMetatable("element")

	L.SetFuncs(-1, map[string]lua.GFunction{
		"__call": func(L *lua.State) int {
			e := CheckElement(L)
			query := L.CheckString(2)
			e.Select(L, query)
			return 1
		},
		"__tostring": func(L *lua.State) int {
			L.PushString(fmt.Sprintf("%s", CheckElement(L).name))
			return 1
		},
	})

	L.CreateTable(0, 0)
	{
		L.SetFuncs(-1, map[string]lua.GFunction{
			"all": func(L *lua.State) int {
				e := CheckElement(L)
				query := L.CheckString(2)
				e.SelectAll(L, query)
				return 1
			},
			"sendKeys":   fn(Element.SendKeys),
			"setValue":   fn(Element.SetValue),
			"click":      fn(Element.Click),
			"submit":     fn(Element.Submit),
			"focus":      fn(Element.Focus),
			"blur":       fn(Element.Blur),
			"screenshot": fn(Element.Screenshot),
		})

		L.CreateTable(0, 0)
		{
			L.PushFunction(func(L *lua.State) int {
				e := CheckElement(L)
				switch L.CheckString(2) {
				case "text":
					return e.GetText(L)
				case "innerHTML":
					return e.GetInnerHTML(L)
				case "outerHTML":
					return e.GetOuterHTML(L)
				case "value":
					return e.GetValue(L)
				default:
					return e.GetAttribute(L)
				}
			})
			L.SetField(-2, "__index")
		}
		L.SetMetatable(-2)
	}
	L.SetField(-2, "__index")

	L.SetGlobal("element")
}
