package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/yuin/gopher-lua"
)

type LoadWaiter struct {
	sync.Mutex

	waits map[network.RequestID]chan struct{}
}

func NewLoadWaiter() *LoadWaiter {
	return &LoadWaiter{
		waits: make(map[network.RequestID]chan struct{}),
	}
}

func (l *LoadWaiter) Complete(id network.RequestID) {
	l.Lock()
	defer l.Unlock()
	if ch, ok := l.waits[id]; ok {
		close(ch)
		delete(l.waits, id)
	}
}

func (l *LoadWaiter) Wait(id network.RequestID) {
	ch := make(chan struct{})
	l.Lock()
	l.waits[id] = ch
	l.Unlock()
	<-ch
}

type Tab struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	env    *Environment

	loading *LoadWaiter

	width, height int64
	onDialog      *lua.LFunction
	onDownloaded  *lua.LFunction
	onRequest     *lua.LFunction
	onResponse    *lua.LFunction

	recorder *Recorder
}

func NewTab(ctx context.Context, env *Environment, url string) *Tab {
	t := AsyncRun(env, func() *Tab {
		ctx, cancel := chromedp.NewContext(ctx)
		t := &Tab{
			ctx:     ctx,
			cancel:  cancel,
			env:     env,
			loading: NewLoadWaiter(),
			width:   1280,
			height:  720,
		}
		t.runInCallback(
			browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).WithDownloadPath(env.storage.Dir).WithEventsEnabled(true),
			chromedp.EmulateViewport(t.width, t.height),
		)
		return t
	})

	if env.EnableRecording {
		var err error
		t.recorder, err = NewRecorder(ctx)
		if err != nil {
			env.RaiseError("%s", err)
		}
	}

	if url != "" {
		t.Run(fmt.Sprintf("$:go(%q)", url), chromedp.Navigate(url))
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
					t.wg.Add(1)
					go func() {
						t.env.CallEventHandler(
							t.onDownloaded,
							map[string]lua.LValue{
								"filepath": lua.LString(filepath),
								"bytes":    lua.LNumber(e.TotalBytes),
							},
							0,
						)
						t.wg.Done()
					}()
				}
			case browser.DownloadProgressStateCanceled:
				t.env.storage.CancelDownload(e.GUID)
			}
		case *page.EventJavascriptDialogOpening:
			t.wg.Add(1)
			go func() {
				if t.onDialog == nil {
					t.runInCallback(page.HandleJavaScriptDialog(true))
				} else {
					result := t.env.CallEventHandler(
						t.onDialog,
						map[string]lua.LValue{
							"type":    lua.LString(e.Type),
							"message": lua.LString(e.Message),
							"url":     lua.LString(e.URL),
						},
						2,
					)

					action := page.HandleJavaScriptDialog(lua.LVAsBool(result[0]))

					if result[1].Type() != lua.LTNil {
						action = action.WithPromptText(string(lua.LVAsString(result[1])))
					}
					t.runInCallback(action)
				}
				t.wg.Done()
			}()
		case *network.EventRequestWillBeSent:
			if t.onRequest != nil {
				t.wg.Add(1)
				go func() {
					params := map[string]lua.LValue{
						"id":     lua.LString(e.RequestID.String()),
						"type":   lua.LString(e.Type.String()),
						"url":    lua.LString(e.DocumentURL),
						"method": lua.LString(e.Request.Method),
					}
					if e.Request.HasPostData {
						params["body"] = lua.LString(e.Request.PostData)
					} else {
						params["body"] = lua.LNil
					}
					t.env.CallEventHandler(t.onRequest, params, 0)
					t.wg.Done()
				}()
			}
		case *network.EventLoadingFinished:
			t.loading.Complete(e.RequestID)
		case *network.EventResponseReceived:
			if t.onResponse != nil {
				t.wg.Add(1)
				go func() {
					params := map[string]lua.LValue{
						"id":         lua.LString(e.RequestID.String()),
						"type":       lua.LString(e.Type.String()),
						"url":        lua.LString(e.Response.URL),
						"status":     lua.LNumber(e.Response.Status),
						"mimetype":   lua.LString(e.Response.MimeType),
						"remoteIP":   lua.LString(e.Response.RemoteIPAddress),
						"remotePort": lua.LNumber(e.Response.RemotePort),
						"length":     lua.LNumber(e.Response.EncodedDataLength),
						"body": t.env.NewFunction(func(L *lua.LState) int {
							id, ok := L.GetField(L.CheckTable(1), "id").(lua.LString)
							if !ok {
								L.ArgError(1, "expected a table contains id.")
							}

							var body []byte
							t.Run("", chromedp.ActionFunc(func(ctx context.Context) (err error) {
								t.loading.Wait(network.RequestID(id))
								body, err = network.GetResponseBody(network.RequestID(id)).Do(ctx)
								var cdperr *cdproto.Error
								if errors.As(err, &cdperr) && cdperr.Code == -32000 {
									// -32000 means "no data found"
									body = nil
									err = nil
								}
								return err
							}))
							if body == nil {
								L.Push(lua.LNil)
							} else {
								L.Push(lua.LString(string(body)))
							}
							return 1
						}),
					}
					t.env.CallEventHandler(t.onResponse, params, 0)
					t.wg.Done()
				}()
			}
		}
	})

	return lt
}

func (t *Tab) RecordOnce(L *lua.LState, taskName string, isAfter bool) {
	if taskName != "" && t.recorder != nil {
		var buf []byte
		t.Run(
			"",
			chromedp.CaptureScreenshot(&buf),
			chromedp.ActionFunc(func(ctx context.Context) error {
				return t.recorder.RecordOnce(taskName, isAfter, buf)
			}),
		)
	}
}

func (t *Tab) handleError(err error) {
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			t.env.RaiseError("timeout")
		} else {
			t.env.RaiseError("%s", err)
		}
	}
}

// runCallback execute browser action without handle GIL.
func (t *Tab) runInCallback(actions ...chromedp.Action) {
	t.handleError(chromedp.Run(t.ctx, actions...))
}

func (t *Tab) Run(taskName string, action ...chromedp.Action) {
	err := AsyncRun(t.env, func() error {
		if taskName != "" && t.recorder != nil {
			var before, after []byte

			action = append(
				[]chromedp.Action{chromedp.CaptureScreenshot(&before)},
				action...,
			)
			action = append(
				action,
				chromedp.CaptureScreenshot(&after),
				chromedp.ActionFunc(func(ctx context.Context) error {
					return t.recorder.RecordBoth(taskName, before, after)
				}),
			)
		}
		return chromedp.Run(t.ctx, action...)
	})
	t.handleError(err)
}

func (t *Tab) RunSelector(query string, action ...chromedp.Action) {
	err := AsyncRun(t.env, func() error {
		ctx, cancel := context.WithTimeout(t.ctx, time.Second)
		defer cancel()

		return chromedp.Run(ctx, action...)
	})
	t.handleError(err)
}

func (t *Tab) Save(name, ext string, data []byte) error {
	return t.env.storage.Save(name, ext, data)
}

func (t *Tab) Go(L *lua.LState) {
	url := L.CheckString(2)

	t.Run(fmt.Sprintf("$:go(%q)", url), chromedp.Navigate(url))
}

func (t *Tab) Forward(L *lua.LState) {
	t.Run("$:forward()", chromedp.NavigateForward())
}

func (t *Tab) Back(L *lua.LState) {
	t.Run("$:back()", chromedp.NavigateBack())
}

func (t *Tab) Reload(L *lua.LState) {
	t.Run("$:reload()", chromedp.Reload())
}

func (t *Tab) Close(L *lua.LState) {
	t.wg.Wait()

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
	t.Run(
		"",
		chromedp.CaptureScreenshot(&buf),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return t.Save(name, ".jpg", buf)
		}),
	)
}

func (t *Tab) SetViewport(L *lua.LState) {
	w := L.CheckInt64(2)
	h := L.CheckInt64(3)

	t.Run(fmt.Sprintf("$:setViewport(%d, %d)", w, h), chromedp.EmulateViewport(w, h))
	t.width, t.height = w, h
}

func (t *Tab) Wait(L *lua.LState) {
	query := L.CheckString(2)

	t.Run(fmt.Sprintf("$:wait(%q)", query), chromedp.WaitVisible(query, chromedp.ByQuery))
}

func (t *Tab) OnDialog(L *lua.LState) {
	t.onDialog = L.OptFunction(2, nil)
}

func (t *Tab) OnDownloaded(L *lua.LState) {
	t.onDownloaded = L.OptFunction(2, nil)
}

func (t *Tab) updateNetworkConfig() {
	if t.onRequest == nil && t.onResponse == nil {
		t.Run("", network.Disable())
	} else {
		t.Run("", network.Enable())
	}
}

func (t *Tab) OnRequest(L *lua.LState) {
	t.onRequest = L.OptFunction(2, nil)
	t.updateNetworkConfig()
}

func (t *Tab) OnResponse(L *lua.LState) {
	t.onResponse = L.OptFunction(2, nil)
	t.updateNetworkConfig()
}

func (t *Tab) Eval(L *lua.LState) int {
	script := L.CheckString(2)

	var res any
	t.Run(fmt.Sprintf("$:eval([[ %s ]])", script), chromedp.Evaluate(script, &res))
	L.Push(PackLValue(L, res))
	return 1
}

func (t *Tab) GetURL(L *lua.LState) int {
	var url string
	t.Run("", chromedp.Location(&url))
	L.Push(lua.LString(url))
	return 1
}

func (t *Tab) GetTitle(L *lua.LState) int {
	var title string
	t.Run("", chromedp.Title(&title))
	L.Push(lua.LString(title))
	return 1
}

func (t *Tab) GetViewport(L *lua.LState) int {
	t.env.Yield()

	v := L.NewTable()
	L.SetField(v, "width", lua.LNumber(t.width))
	L.SetField(v, "height", lua.LNumber(t.height))
	L.Push(v)
	return 1
}

func RegisterTabType(ctx context.Context, env *Environment) {
	fn := func(f func(*Tab, *lua.LState)) *lua.LFunction {
		return env.NewFunction(func(L *lua.LState) int {
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
		"onRequest":    fn((*Tab).OnRequest),
		"onResponse":   fn((*Tab).OnResponse),
		"all": env.NewFunction(func(L *lua.LState) int {
			L.Push(NewElementsArray(L, CheckTab(L), L.CheckString(2)).ToLua(L))
			return 1
		}),
		"eval": env.NewFunction(func(L *lua.LState) int {
			return CheckTab(L).Eval(L)
		}),
	}

	getters := map[string]func(*Tab, *lua.LState) int{
		"url":      (*Tab).GetURL,
		"title":    (*Tab).GetTitle,
		"viewport": (*Tab).GetViewport,
	}

	env.RegisterNewType("tab", map[string]lua.LGFunction{
		"new": func(L *lua.LState) int {
			var url string
			if L.GetTop() > 0 {
				url = L.CheckString(1)
			}

			t := NewTab(ctx, env, url)
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
			t.Run("", chromedp.Location(&url), chromedp.Title(&title))
			L.Push(lua.LString("[" + title + "](" + url + ")"))
			return 1
		},
	}, nil)
}
