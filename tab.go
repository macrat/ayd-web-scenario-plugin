package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/yuin/gopher-lua"
)

type Tab struct {
	ctx    context.Context
	cancel context.CancelFunc

	width, height int64
}

func NewTab(ctx context.Context, L *lua.LState) *Tab {
	ctx, cancel := chromedp.NewContext(ctx)
	t := &Tab{
		ctx:    ctx,
		cancel: cancel,
		width:  1280,
		height: 720,
	}

	t.Run(L, chromedp.EmulateViewport(t.width, t.height))

	if L.GetTop() > 0 {
		url := L.CheckString(1)
		t.Run(L, chromedp.Navigate(url))
	}

	return t
}

func CheckTab(L *lua.LState) *Tab {
	if ud, ok := L.Get(1).(*lua.LUserData); ok {
		if t, ok := ud.Value.(*Tab); ok {
			return t
		}
	}

	L.ArgError(1, "tab expected. perhaps you call it like tab.xxx() instead of tab:xxx().")
	return nil
}

func (t *Tab) Run(L *lua.LState, action ...chromedp.Action) {
	if err := chromedp.Run(t.ctx, action...); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			L.RaiseError("timeout")
		} else {
			L.RaiseError("%s", err)
		}
	}
}

func (t *Tab) RunSelector(L *lua.LState, query string, action ...chromedp.Action) {
	ctx, cancel := context.WithTimeout(t.ctx, time.Second)
	defer cancel()

	if err := chromedp.Run(ctx, action...); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			L.RaiseError("no such element: %s", query)
		} else {
			L.RaiseError("%s", err)
		}
	}
}

func (t *Tab) Go(L *lua.LState) {
	url := L.CheckString(2)
	t.Run(L, chromedp.Navigate(url))
}

func (t *Tab) Forward(L *lua.LState) {
	t.Run(L, chromedp.NavigateForward())
}

func (t *Tab) Back(L *lua.LState) {
	t.Run(L, chromedp.NavigateBack())
}

func (t *Tab) Reload(L *lua.LState) {
	t.Run(L, chromedp.Reload())
}

func (t *Tab) Close(L *lua.LState) {
	t.cancel()
}

func (t *Tab) Screenshot(L *lua.LState) {
	name := L.ToString(2)

	var buf []byte
	t.Run(L, chromedp.CaptureScreenshot(&buf))
	os.WriteFile(name+".jpg", buf, 0644)
}

func (t *Tab) SetViewport(L *lua.LState) {
	w := L.CheckInt64(2)
	h := L.CheckInt64(3)

	// don't modify property of Tab before check both of arguments.
	t.Run(L, chromedp.EmulateViewport(w, h))
	t.width, t.height = w, h
}

func (t *Tab) Wait(L *lua.LState) {
	query := L.CheckString(2)

	t.Run(L, chromedp.WaitVisible(query, chromedp.ByQuery))
}

func (t *Tab) GetURL(L *lua.LState) int {
	var url string
	t.Run(L, chromedp.Location(&url))
	L.Push(lua.LString(url))
	return 1
}

func (t *Tab) GetTitle(L *lua.LState) int {
	var title string
	t.Run(L, chromedp.Title(&title))
	L.Push(lua.LString(title))
	return 1
}

func (t *Tab) GetViewport(L *lua.LState) int {
	v := L.NewTable()
	L.SetField(v, "width", lua.LNumber(t.width))
	L.SetField(v, "height", lua.LNumber(t.height))
	L.Push(v)
	return 1
}

func RegisterTabType(ctx context.Context, L *lua.LState) {
	fn := func(f func(*Tab, *lua.LState)) *lua.LFunction {
		return L.NewFunction(func(L *lua.LState) int {
			f(CheckTab(L), L)
			L.Push(L.Get(1))
			return 1
		})
	}

	methods := map[string]*lua.LFunction{
		"go":          fn((*Tab).Go),
		"forward":     fn((*Tab).Forward),
		"back":        fn((*Tab).Back),
		"reload":      fn((*Tab).Reload),
		"close":       fn((*Tab).Close),
		"screenshot":  fn((*Tab).Screenshot),
		"setViewport": fn((*Tab).SetViewport),
		"wait":        fn((*Tab).Wait),
		"all": L.NewFunction(func(L *lua.LState) int {
			L.Push(NewElementsArray(L, CheckTab(L), L.CheckString(2)).ToLua(L))
			return 1
		}),
	}

	getters := map[string]func(*Tab, *lua.LState) int{
		"url":      (*Tab).GetURL,
		"title":    (*Tab).GetTitle,
		"viewport": (*Tab).GetViewport,
	}

	tab := L.SetFuncs(L.NewTypeMetatable("tab"), map[string]lua.LGFunction{
		"new": func(L *lua.LState) int {
			t := L.NewUserData()
			t.Value = NewTab(ctx, L)
			L.SetMetatable(t, L.GetTypeMetatable("tab"))
			L.Push(t)
			return 1
		},
		"__call": func(L *lua.LState) int {
			L.Push(NewElement(L, CheckTab(L), L.CheckString(2)).ToLua(L))
			return 1
		},
		"__index": func(L *lua.LState) int {
			name := L.CheckString(2)

			if f, ok := getters[name]; ok {
				return f(CheckTab(L), L)
			} else if f, ok := methods[name]; ok {
				L.Push(f)
				return 1
			} else {
				return 0
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
