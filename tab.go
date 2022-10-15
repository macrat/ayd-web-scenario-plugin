package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/yuin/gopher-lua"
)

type Tab struct {
	ctx    context.Context
	cancel context.CancelFunc
	env    *Environment

	width, height int64
	onDialog      *lua.LFunction
	onDownloaded  *lua.LFunction

	recorder *Recorder
}

func NewTab(ctx context.Context, env *Environment) *Tab {
	ctx, cancel := chromedp.NewContext(ctx)
	t := &Tab{
		ctx:    ctx,
		cancel: cancel,
		env:    env,
		width:  1280,
		height: 720,
	}

	if env.EnableRecording {
		var err error
		t.recorder, err = NewRecorder(ctx)
		if err != nil {
			env.L.RaiseError("%s", err)
		}
	}

	t.Run(
		env.L,
		"",
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).WithDownloadPath(env.storage.Dir).WithEventsEnabled(true),
		chromedp.EmulateViewport(t.width, t.height),
	)

	if env.L.GetTop() > 0 {
		url := env.L.CheckString(1)
		t.Run(env.L, fmt.Sprintf("$:go(%q)", url), chromedp.Navigate(url))
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

func (t *Tab) ToLua(L *lua.LState) *lua.LUserData {
	lt := L.NewUserData()
	lt.Value = t
	L.SetMetatable(lt, L.GetTypeMetatable("tab"))

	chromedp.ListenTarget(t.ctx, func(ev any) {
		switch e := ev.(type) {
		case *browser.EventDownloadWillBegin:
			t.env.storage.StartDownload(e.GUID, e.SuggestedFilename)
		case *browser.EventDownloadProgress:
			switch e.State {
			case browser.DownloadProgressStateCompleted:
				filepath := t.env.storage.CompleteDownload(e.GUID)
				if t.onDownloaded != nil {
					L.Push(t.onDownloaded)
					L.Push(lua.LString(filepath))
					L.Push(lua.LNumber(e.TotalBytes))
					L.Call(2, 0)
				}
			case browser.DownloadProgressStateCanceled:
				t.env.storage.CancelDownload(e.GUID)
			}
		case *page.EventJavascriptDialogOpening:
			if t.onDialog == nil {
				page.HandleJavaScriptDialog(true)
			} else {
				L.Push(t.onDialog)
				L.Push(lua.LString(e.Type))
				L.Push(lua.LString(e.Message))
				L.Push(lua.LString(e.URL))
				L.Call(3, 2)

				action := page.HandleJavaScriptDialog(L.ToBool(-2))

				if L.Get(-1).Type() != lua.LTNil {
					action = action.WithPromptText(string(L.ToString(-1)))
				}

				go func() {
					t.Run(L, "", action)
				}()
			}
		}
	})

	return lt
}

func (t *Tab) RecordOnce(L *lua.LState, taskName string, isAfter bool) {
	if taskName != "" && t.recorder != nil {
		var buf []byte
		err := chromedp.Run(t.ctx, chromedp.CaptureScreenshot(&buf))
		t.recorder.RecordOnce(taskName, isAfter, buf)
		if err != nil {
			L.RaiseError("%s", err)
		}
	}
}

func (t *Tab) Run(L *lua.LState, taskName string, action ...chromedp.Action) {
	var before, after []byte
	var err error

	if taskName != "" && t.recorder != nil {
		err = chromedp.Run(t.ctx, append(
			[]chromedp.Action{chromedp.CaptureScreenshot(&before)},
			append(action, chromedp.CaptureScreenshot(&after))...,
		)...)
		t.recorder.RecordBoth(taskName, before, after)
	} else {
		err = chromedp.Run(t.ctx, action...)
	}

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			L.RaiseError("timeout")
		} else {
			L.RaiseError("%s", err)
		}
	}
}

func (t *Tab) Save(L *lua.LState, name, ext string, data []byte) {
	err := t.env.storage.Save(name, ext, data)
	if err != nil {
		L.RaiseError("%s", err)
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

	t.Run(L, fmt.Sprintf("$:go(%q)", url), chromedp.Navigate(url))
}

func (t *Tab) Forward(L *lua.LState) {
	t.Run(L, "$:forward()", chromedp.NavigateForward())
}

func (t *Tab) Back(L *lua.LState) {
	t.Run(L, "$:back()", chromedp.NavigateBack())
}

func (t *Tab) Reload(L *lua.LState) {
	t.Run(L, "$:reload()", chromedp.Reload())
}

func (t *Tab) Close(L *lua.LState) {
	if t.recorder != nil {
		t.RecordOnce(L, "$:close()", false)
	}

	t.env.unregisterTab(t)

	t.cancel()

	if t.recorder != nil {
		t.env.saveRecord(t.recorder)
	}
}

func (t *Tab) Screenshot(L *lua.LState) {
	name := L.ToString(2)

	var buf []byte
	t.Run(L, "", chromedp.CaptureScreenshot(&buf))
	t.Save(L, name, ".jpg", buf)
}

func (t *Tab) SetViewport(L *lua.LState) {
	w := L.CheckInt64(2)
	h := L.CheckInt64(3)

	t.Run(L, fmt.Sprintf("$:setViewport(%d, %d)", w, h), chromedp.EmulateViewport(w, h))
	t.width, t.height = w, h
}

func (t *Tab) Wait(L *lua.LState) {
	query := L.CheckString(2)

	t.Run(L, fmt.Sprintf("$:wait(%q)", query), chromedp.WaitVisible(query, chromedp.ByQuery))
}

func (t *Tab) OnDialog(L *lua.LState) {
	t.onDialog = L.CheckFunction(2)
}

func (t *Tab) OnDownloaded(L *lua.LState) {
	t.onDownloaded = L.CheckFunction(2)
}

func (t *Tab) Eval(L *lua.LState) int {
	script := L.CheckString(2)

	var res any
	t.Run(L, fmt.Sprintf("$:eval([[ %s ]])", script), chromedp.Evaluate(script, &res))
	L.Push(PackLValue(L, res))
	return 1
}

func (t *Tab) GetURL(L *lua.LState) int {
	var url string
	t.Run(L, "", chromedp.Location(&url))
	L.Push(lua.LString(url))
	return 1
}

func (t *Tab) GetTitle(L *lua.LState) int {
	var title string
	t.Run(L, "", chromedp.Title(&title))
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

func RegisterTabType(ctx context.Context, env *Environment) {
	fn := func(f func(*Tab, *lua.LState)) *lua.LFunction {
		return env.L.NewFunction(func(L *lua.LState) int {
			f(CheckTab(L), L)
			L.Push(L.Get(1))
			return 1
		})
	}

	methods := map[string]*lua.LFunction{
		"go":           fn((*Tab).Go),
		"forward":      fn((*Tab).Forward),
		"back":         fn((*Tab).Back),
		"reload":       fn((*Tab).Reload),
		"close":        fn((*Tab).Close),
		"screenshot":   fn((*Tab).Screenshot),
		"setViewport":  fn((*Tab).SetViewport),
		"wait":         fn((*Tab).Wait),
		"onDialog":     fn((*Tab).OnDialog),
		"onDownloaded": fn((*Tab).OnDownloaded),
		"all": env.L.NewFunction(func(L *lua.LState) int {
			L.Push(NewElementsArray(L, CheckTab(L), L.CheckString(2)).ToLua(L))
			return 1
		}),
		"eval": env.L.NewFunction(func(L *lua.LState) int {
			return CheckTab(L).Eval(L)
		}),
	}

	getters := map[string]func(*Tab, *lua.LState) int{
		"url":      (*Tab).GetURL,
		"title":    (*Tab).GetTitle,
		"viewport": (*Tab).GetViewport,
	}

	tab := env.L.SetFuncs(env.L.NewTypeMetatable("tab"), map[string]lua.LGFunction{
		"new": func(L *lua.LState) int {
			t := NewTab(ctx, env)
			env.registerTab(t)
			L.Push(t.ToLua(L))
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
			t.Run(L, "", chromedp.Location(&url), chromedp.Title(&title))
			L.Push(lua.LString("[" + title + "](" + url + ")"))
			return 1
		},
	})

	env.L.SetGlobal("tab", tab)
}
