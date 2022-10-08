package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chromedp/chromedp"
	"github.com/yuin/gopher-lua"
)

type Tab struct {
	ctx    context.Context
	cancel context.CancelFunc

	width, height int64
}

func CheckTab(L *lua.LState) *Tab {
	if t, ok := L.CheckUserData(1).Value.(*Tab); ok {
		return t
	}
	L.ArgError(1, "tab expected. perhaps you call it like tab.xxx() instead of tab:xxx().")
	return nil
}

func (t *Tab) Run(L *lua.LState, action ...chromedp.Action) {
	if err := chromedp.Run(t.ctx, action...); err != nil {
		L.RaiseError("%s", err)
	}
}

func RegisterTabType(ctx context.Context, L *lua.LState) {
	methods := map[string]*lua.LFunction{
		"go": L.NewFunction(func(L *lua.LState) int {
			url := L.CheckString(2)
			fmt.Printf("navigate goto %s\n", url)
			CheckTab(L).Run(L, chromedp.Navigate(url))

			L.Push(L.Get(1))
			return 1
		}),
		"forward": L.NewFunction(func(L *lua.LState) int {
			fmt.Printf("navigate go forward\n")
			CheckTab(L).Run(L, chromedp.NavigateForward())

			L.Push(L.Get(1))
			return 1
		}),
		"back": L.NewFunction(func(L *lua.LState) int {
			fmt.Printf("navigate go back\n")
			CheckTab(L).Run(L, chromedp.NavigateBack())

			L.Push(L.Get(1))
			return 1
		}),
		"reload": L.NewFunction(func(L *lua.LState) int {
			fmt.Printf("reload page\n")
			CheckTab(L).Run(L, chromedp.Reload())

			L.Push(L.Get(1))
			return 1
		}),
		"close": L.NewFunction(func(L *lua.LState) int {
			fmt.Printf("close tab\n")
			CheckTab(L).cancel()

			L.Push(L.Get(1))
			return 1
		}),
		"screenshot": L.NewFunction(func(L *lua.LState) int {
			var buf []byte
			name := L.CheckString(2)
			CheckTab(L).Run(L, chromedp.CaptureScreenshot(&buf))
			os.WriteFile(name, buf, 0644)

			L.Push(L.Get(1))
			return 1
		}),
		"wait": L.NewFunction(func(L *lua.LState) int {
			fmt.Printf("wait for document ready\n")
			CheckTab(L).Run(L, chromedp.WaitReady("document"))

			L.Push(L.Get(1))
			return 0
		}),
		"setViewport": L.NewFunction(func(L *lua.LState) int {
			tab := CheckTab(L)
			tab.width = L.CheckInt64(2)
			tab.height = L.CheckInt64(3)
			fmt.Printf("change viewport to %dx%d\n", tab.width, tab.height)
			tab.Run(L, chromedp.EmulateViewport(tab.width, tab.height))

			L.Push(L.Get(1))
			return 0
		}),
		"all": L.NewFunction(func(L *lua.LState) int {
			t := CheckTab(L)
			L.Push(NewElementArray(L, t, L.CheckString(2)))
			return 1
		}),
	}

	tab := L.SetFuncs(L.NewTypeMetatable("tab"), map[string]lua.LGFunction{
		"new": func(L *lua.LState) int {
			ctx, cancel := chromedp.NewContext(ctx)
			tab := &Tab{
				ctx:    ctx,
				cancel: cancel,
				width:  1280,
				height: 720,
			}

			tab.Run(L, chromedp.EmulateViewport(tab.width, tab.height))

			if L.GetTop() == 0 {
				fmt.Printf("open new blank tab\n")
			} else {
				url := L.CheckString(1)
				fmt.Printf("open %s on new tab\n", url)
				tab.Run(L, chromedp.Navigate(url))
			}

			t := L.NewUserData()
			t.Value = tab
			L.SetMetatable(t, L.GetTypeMetatable("tab"))
			L.Push(t)
			return 1
		},
		"__call": func(L *lua.LState) int {
			t := CheckTab(L)
			L.Push(NewElement(L, t, L.CheckString(2)).ToLua(L))
			return 1
		},
		"__index": func(L *lua.LState) int {
			name := L.CheckString(2)

			switch name {
			case "url":
				var url string
				CheckTab(L).Run(L, chromedp.Location(&url))
				L.Push(lua.LString(url))
				return 1
			case "title":
				var title string
				CheckTab(L).Run(L, chromedp.Title(&title))
				L.Push(lua.LString(title))
				return 1
			case "viewport":
				t := CheckTab(L)
				v := L.NewTable()
				L.SetField(v, "width", lua.LNumber(t.width))
				L.SetField(v, "height", lua.LNumber(t.height))
				L.Push(v)
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
			var url, title string
			t := CheckTab(L)
			t.Run(L, chromedp.Location(&url), chromedp.Title(&title))
			L.Push(lua.LString("[" + title + "](" + url + ")"))
			return 1
		},
	})

	L.SetGlobal("tab", tab)
}
