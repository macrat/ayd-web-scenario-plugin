package webscenario

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	//"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"github.com/macrat/ayd-web-scenario/internal/lua"
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

	id            int
	width, height int64
	//dialogEvent   *EventHandler
	//downloadEvent *EventHandler
	//requestEvent  *EventHandler
	//responseEvent *EventHandler

	recorder *Recorder
}

func NewTab(ctx context.Context, L *lua.State, env *Environment, id int) *Tab {
	url := ""
	width, height := int64(800), int64(800)
	userAgent := ""
	recording := false

	switch L.Type(1) {
	case lua.String:
		url = L.ToString(1)
	case lua.Table:
		L.GetField(1, "url")
		if L.Type(-1) != lua.Nil {
			url = L.ToString(-1)
		}
		L.Pop(1)

		L.GetField(1, "width")
		if L.Type(-1) == lua.Number {
			width = int64(L.ToNumber(-1))
		}
		L.Pop(1)

		L.GetField(1, "height")
		if L.Type(-1) == lua.Number {
			height = int64(L.ToNumber(-1))
		}
		L.Pop(1)

		L.GetField(1, "useragent")
		if L.Type(-1) != lua.Nil {
			userAgent = L.ToString(-1)
		}
		L.Pop(1)

		L.GetField(1, "recording")
		recording = L.ToBoolean(-1)
		L.Pop(1)
	case lua.Nil:
	default:
		L.ArgErrorf(1, "nil, string, or table expected, got %s", L.Type(1))
	}

	t := AsyncRun(env, L, func() (*Tab, error) {
		ctx, cancel := chromedp.NewContext(ctx)
		t := &Tab{
			ctx:     ctx,
			cancel:  cancel,
			env:     env,
			loading: NewLoadWaiter(),

			id:     id,
			width:  width,
			height: height,

			//dialogEvent:   NewEventHandler((*Tab).HandleDialog),
			//downloadEvent: NewEventHandler((*Tab).HandleEvent),
			//requestEvent:  NewEventHandler((*Tab).HandleEvent),
			//responseEvent: NewEventHandler((*Tab).HandleEvent),
		}
		err := t.RunInCallback(
			browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).WithDownloadPath(env.storage.Dir).WithEventsEnabled(true),
			chromedp.Emulate(device.Info{
				UserAgent: userAgent,
				Width:     t.width,
				Height:    t.height,
				Scale:     1,
			}),
		)
		return t, err
	})

	if recording || env.EnableRecording {
		t.recorder = NewRecorder(t.ctx, int(width), int(height))
	}

	if url != "" {
		t.Run(L, fmt.Sprintf("$:go(%q)", url), true, 0, chromedp.Navigate(url))
	}

	return t
}

func CheckTab(L *lua.State) *Tab {
	if t, ok := L.ToUserdata(1).(*Tab); ok {
		return t
	}

	L.ArgErrorf(1, "tab expected, got %s. perhaps you call it like tab.xxx() instead of tab:xxx().", L.Type(1))
	return nil
}

func (t *Tab) PushTo(L *lua.State) {
	L.PushUserdata(t)
	L.GetTypeMetatable("tab")
	L.SetMetatable(-2)

	/*
		chromedp.ListenTarget(t.ctx, func(ev any) {
			switch e := ev.(type) {
			case *page.EventJavascriptDialogOpening:
				ev := t.env.BuildTable(func(L *lua.LState, ev *lua.LTable) {
					L.SetField(ev, "type", lua.LString(e.Type.String()))
					L.SetField(ev, "message", lua.LString(e.Message))
					L.SetField(ev, "url", lua.LString(e.URL))
				})
				t.dialogEvent.Invoke(t, ev)
			case *browser.EventDownloadWillBegin:
				t.env.storage.StartDownload(e.GUID, e.SuggestedFilename)
			case *browser.EventDownloadProgress:
				switch e.State {
				case browser.DownloadProgressStateCompleted:
					ev := t.env.BuildTable(func(L *lua.LState, ev *lua.LTable) {
						L.SetField(ev, "path", lua.LString(t.env.storage.CompleteDownload(e.GUID)))
						L.SetField(ev, "bytes", lua.LNumber(e.TotalBytes))
					})
					t.downloadEvent.Invoke(t, ev)
				case browser.DownloadProgressStateCanceled:
					t.env.storage.CancelDownload(e.GUID)
				}
			case *network.EventRequestWillBeSent:
				ev := t.env.BuildTable(func(L *lua.LState, ev *lua.LTable) {
					L.SetField(ev, "id", lua.LString(e.RequestID.String()))
					L.SetField(ev, "type", lua.LString(e.Type.String()))
					L.SetField(ev, "url", lua.LString(e.DocumentURL))
					L.SetField(ev, "method", lua.LString(e.Request.Method))
					L.SetField(ev, "headers", PushAny(L, e.Request.Headers))
					if e.Request.HasPostData {
						L.SetField(ev, "body", lua.LString(e.Request.PostData))
					}
				})
				t.requestEvent.Invoke(t, ev)
			case *network.EventLoadingFinished:
				t.loading.Complete(e.RequestID)
			case *network.EventResponseReceived:
				ev := t.env.BuildTable(func(L *lua.LState, ev *lua.LTable) {
					L.SetField(ev, "id", lua.LString(e.RequestID.String()))
					L.SetField(ev, "type", lua.LString(e.Type.String()))
					L.SetField(ev, "url", lua.LString(e.Response.URL))
					L.SetField(ev, "status", lua.LNumber(e.Response.Status))
					L.SetField(ev, "headers", PushAny(L, e.Response.Headers))
					L.SetField(ev, "length", lua.LNumber(e.Response.EncodedDataLength))
					L.SetField(ev, "remoteIP", lua.LString(e.Response.RemoteIPAddress))
					L.SetField(ev, "remotePort", lua.LNumber(e.Response.RemotePort))

					L.SetMetatable(ev, AsFileLikeMeta(L, NewDelayedReader(func() io.Reader {
						var body []byte
						t.Run(L, "$response:read()", false, 0, chromedp.ActionFunc(func(ctx context.Context) (err error) {
							ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
							defer cancel()

							t.loading.Wait(e.RequestID)
							body, err = network.GetResponseBody(e.RequestID).Do(ctx)
							var cdperr *cdproto.Error
							if errors.As(err, &cdperr) && cdperr.Code == -32000 {
								// -32000 means "no data found"
								body = nil
								err = nil
							}
							return err
						}))
						return bytes.NewReader(body)
					})))
				})

				t.responseEvent.Invoke(t, ev)
			}
		})
	*/
}

func captureScreenshotForRecording(buf *[]byte) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		*buf, err = page.CaptureScreenshot().
			WithCaptureBeyondViewport(true).
			WithFormat(page.CaptureScreenshotFormatPng).
			WithOptimizeForSpeed(true).
			Do(ctx)
		return err
	})
}

func (t *Tab) RecordOnce(L *lua.State, taskName string) {
	if taskName != "" && t.recorder != nil {
		where, line := L.Where(1)

		var buf []byte
		t.Run(
			L,
			taskName,
			false,
			0,
			captureScreenshotForRecording(&buf),
			t.recorder.Record(where, line, &buf),
		)
	}
}

// RunCallback execute browser action without to release GIL.
func (t *Tab) RunInCallback(actions ...chromedp.Action) error {
	return chromedp.Run(t.ctx, actions...)
}

func (t *Tab) Run(L *lua.State, taskName string, capture bool, timeout time.Duration, action ...chromedp.Action) {
	where, line := L.Where(1)
	t.env.StartTask(where, line, taskName)

	AsyncRun(t.env, L, func() (struct{}, error) {
		ctx := t.ctx
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		if capture && t.recorder != nil {
			var buf []byte

			action = append(
				action,
				captureScreenshotForRecording(&buf),
				t.recorder.Record(where, line, &buf),
			)
		}
		return struct{}{}, chromedp.Run(ctx, action...)
	})
}

func (t *Tab) RunSelector(L *lua.State, taskName string, action ...chromedp.Action) {
	where, line := L.Where(1)
	t.env.StartTask(where, line, taskName)

	AsyncRun(t.env, L, func() (struct{}, error) {
		ctx, cancel := context.WithTimeout(t.ctx, time.Second)
		defer cancel()

		err := chromedp.Run(ctx, action...)
		if errors.Is(err, context.DeadlineExceeded) {
			return struct{}{}, errors.New("no such element")
		}
		return struct{}{}, err
	})
}

func (t *Tab) Save(name, ext string, data []byte) error {
	return t.env.storage.Save(name, ext, data)
}

func (t *Tab) Go(L *lua.State) {
	url := L.CheckString(2)

	t.Run(L, fmt.Sprintf("$:go(%q)", url), true, 0, chromedp.Navigate(url))
}

func (t *Tab) Forward(L *lua.State) {
	t.Run(L, "$:forward()", true, 0, chromedp.NavigateForward())
}

func (t *Tab) Back(L *lua.State) {
	t.Run(L, "$:back()", true, 0, chromedp.NavigateBack())
}

func (t *Tab) Reload(L *lua.State) {
	t.Run(L, "$:reload()", true, 0, chromedp.Reload())
}

func (t *Tab) Close() error {
	AsyncRun(t.env, t.env.lua, func() (struct{}, error) {
		t.wg.Wait()

		t.env.unregisterTab(t)

		t.cancel()

		if t.recorder != nil {
			t.env.saveRecord(t.id, t.recorder)
		}

		//t.dialogEvent.Close()
		//t.downloadEvent.Close()
		//t.requestEvent.Close()
		//t.responseEvent.Close()

		return struct{}{}, nil
	})
	return nil
}

func (t *Tab) LClose(L *lua.State) {
	if t.recorder != nil {
		t.RecordOnce(L, "$:close()")
	}
	t.Close()
}

func (t *Tab) Screenshot(L *lua.State) {
	name := L.ToString(2)

	var buf []byte
	t.Run(
		L,
		fmt.Sprintf("$:screenshot(%v)", name),
		false,
		0,
		chromedp.CaptureScreenshot(&buf),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return t.Save(name, ".png", buf)
		}),
	)
}

func (t *Tab) Wait(L *lua.State) {
	query := L.CheckString(2)
	timeout := time.Duration(float64(L.ToNumber(3)) * float64(time.Millisecond))

	t.Run(L, fmt.Sprintf("$:wait(%q)", query), true, timeout, chromedp.WaitReady(query, chromedp.ByQuery))
}

func (t *Tab) WaitVisible(L *lua.State) {
	query := L.CheckString(2)
	timeout := time.Duration(float64(L.ToNumber(3)) * float64(time.Millisecond))

	t.Run(L, fmt.Sprintf("$:waitVisible(%q)", query), true, timeout, chromedp.WaitVisible(query, chromedp.ByQuery))
}

func (t *Tab) WaitXPath(L *lua.State) {
	query := L.CheckString(2)
	timeout := time.Duration(float64(L.ToNumber(3)) * float64(time.Millisecond))

	t.Run(L, fmt.Sprintf("$:waitXPath(%q)", query), true, timeout, chromedp.WaitReady(query, chromedp.BySearch))
}

func (t *Tab) WaitXPathVisible(L *lua.State) {
	query := L.CheckString(2)
	timeout := time.Duration(float64(L.ToNumber(3)) * float64(time.Millisecond))

	t.Run(L, fmt.Sprintf("$:waitXPathVisible(%q)", query), true, timeout, chromedp.WaitVisible(query, chromedp.BySearch))
}

/*
func (t *Tab) WaitEvent(L *lua.State, taskName string, h *EventHandler) int {
	timeout := time.Duration(-1)
	if L.Type(2) == lua.Number {
		timeout = time.Duration(float64(L.ToNumber(2)) * float64(time.Millisecond))
	}

	where, line := L.Where(1)
	t.env.StartTask(where, line, taskName)

	v := AsyncRun(t.env, L, func() (*lua.LTable, error) {
		ctx := t.ctx
		var cancel context.CancelFunc
		if timeout >= 0 {
			ctx, cancel = context.WithTimeout(t.ctx, timeout)
			defer cancel()
		}
		v := h.Wait(ctx)
		if v == nil {
			return nil, errors.New("timeout")
		}
		return v, nil
	})

	t.RecordOnce(L, taskName)
	L.Pop(L.GetTop() - 1)
	L.Push(v)
	return 2
}

func (t *Tab) WaitDialog(L *lua.LState) int {
	return t.WaitEvent(L, "t:waitDialog()", t.dialogEvent)
}

func (t *Tab) WaitDownload(L *lua.LState) int {
	return t.WaitEvent(L, "t:waitDownload()", t.downloadEvent)
}

func (t *Tab) WaitRequest(L *lua.LState) int {
	return t.WaitEvent(L, "t:waitRequest()", t.requestEvent)
}

func (t *Tab) WaitResponse(L *lua.LState) int {
	return t.WaitEvent(L, "t:waitResponse()", t.responseEvent)
}

func (t *Tab) GetDialogs(L *lua.LState) int {
	L.Push(t.dialogEvent.Status(L))
	return 1
}

func (t *Tab) GetDownload(L *lua.LState) int {
	L.Push(t.downloadEvent.Status(L))
	return 1
}

func (t *Tab) GetRequest(L *lua.LState) int {
	L.Push(t.requestEvent.Status(L))
	return 1
}

func (t *Tab) GetResponse(L *lua.LState) int {
	L.Push(t.responseEvent.Status(L))
	return 1
}

func (t *Tab) HandleEvent(f *lua.LFunction, ev *lua.LTable) {
	if f != nil {
		t.wg.Add(1)
		go func() {
			t.env.CallEventHandler(f, ev, 0)
			t.wg.Done()
		}()
	}
}

func (t *Tab) HandleDialog(f *lua.LFunction, ev *lua.LTable) {
	t.wg.Add(1)
	go func() {
		if f == nil {
			t.RunInCallback(page.HandleJavaScriptDialog(true))
		} else {
			result := t.env.CallEventHandler(f, ev, 2)

			action := page.HandleJavaScriptDialog(lua.LVAsBool(result[0]))

			if result[1].Type() != lua.LTNil {
				action = action.WithPromptText(string(lua.LVAsString(result[1])))
			}
			t.RunInCallback(action)
		}
		t.wg.Done()
	}()
}

func (t *Tab) OnDialog(L *lua.LState) {
	t.dialogEvent.SetFunc(L.OptFunction(2, nil))
}

func (t *Tab) OnDownload(L *lua.LState) {
	t.downloadEvent.SetFunc(L.OptFunction(2, nil))
}

func (t *Tab) updateNetworkConfig(L *lua.LState, taskName string) {
	if t.requestEvent.IsFuncSet() || t.responseEvent.IsFuncSet() {
		t.Run(L, taskName, false, 0, network.Enable())
	} else {
		t.Run(L, taskName, false, 0, network.Disable())
	}
}

func (t *Tab) OnRequest(L *lua.LState) {
	t.requestEvent.SetFunc(L.OptFunction(2, nil))
	t.updateNetworkConfig(L, "$:onRequest()")
}

func (t *Tab) OnResponse(L *lua.LState) {
	t.responseEvent.SetFunc(L.OptFunction(2, nil))
	t.updateNetworkConfig(L, "$:onResponse()")
}
*/

func (t *Tab) Eval(L *lua.State) int {
	script := L.CheckString(2)

	var res any
	t.Run(L, fmt.Sprintf("$:eval([[ %s ]])", script), true, 0, chromedp.Evaluate(script, &res))
	L.PushAny(res)
	return 1
}

func (t *Tab) GetURL(L *lua.State) int {
	var url string
	t.Run(L, "$.url", false, 0, chromedp.Location(&url))
	L.PushString(url)
	return 1
}

func (t *Tab) GetTitle(L *lua.State) int {
	var title string
	t.Run(L, "$.title", false, 0, chromedp.Title(&title))
	L.PushString(title)
	return 1
}

func (t *Tab) GetViewport(L *lua.State) int {
	t.env.Yield()

	L.CreateTable(0, 2)
	L.SetInteger(-1, "width", t.width)
	L.SetInteger(-1, "height", t.height)
	return 1
}

func RegisterTabType(ctx context.Context, env *Environment, L *lua.State) {
	fn := func(f func(*Tab, *lua.State)) lua.GFunction {
		return func(L *lua.State) int {
			f(CheckTab(L), L)
			L.SetTop(1)
			return 1
		}
	}

	fret := func(f func(*Tab, *lua.State) int) lua.GFunction {
		return func(L *lua.State) int {
			return f(CheckTab(L), L)
		}
	}

	id := 0

	L.NewTypeMetatable("tab")
	L.SetFuncs(-1, map[string]lua.GFunction{
		"new": func(L *lua.State) int {
			id++
			t := NewTab(ctx, L, env, id)
			env.registerTab(t)
			t.PushTo(L)
			return 1
		},
		"__call": func(L *lua.State) int {
			NewElement(L, CheckTab(L), L.CheckString(2)).PushTo(L)
			return 1
		},
		"__tostring": func(L *lua.State) int {
			L.PushString(fmt.Sprintf("tab#%d", CheckTab(L).id))
			return 1
		},
	})

	L.CreateTable(0, 0)
	{
		L.SetFuncs(-1, map[string]lua.GFunction{
			"go":               fn((*Tab).Go),
			"forward":          fn((*Tab).Forward),
			"back":             fn((*Tab).Back),
			"reload":           fn((*Tab).Reload),
			"close":            fn((*Tab).LClose),
			"screenshot":       fn((*Tab).Screenshot),
			"wait":             fn((*Tab).Wait),
			"waitXPath":        fn((*Tab).WaitXPath),
			"waitVisible":      fn((*Tab).WaitVisible),
			"waitXPathVisible": fn((*Tab).WaitXPathVisible),
			/*
				"waitDialog":       fret((*Tab).WaitDialog),
				"waitDownload":     fret((*Tab).WaitDownload),
				"waitRequest":      fret((*Tab).WaitRequest),
				"waitResponse":     fret((*Tab).WaitResponse),
				"onDialog":         fn((*Tab).OnDialog),
				"onDownload":       fn((*Tab).OnDownload),
				"onRequest":        fn((*Tab).OnRequest),
				"onResponse":       fn((*Tab).OnResponse),
			*/
			"all": func(L *lua.State) int {
				t := CheckTab(L)
				query := L.CheckString(2)
				PushElementsTable(L, t, query)
				return 1
			},
			"xpath": func(L *lua.State) int {
				t := CheckTab(L)
				query := L.CheckString(2)
				PushElementsTableByXPath(L, t, query)
				return 1
			},
			"eval": fret((*Tab).Eval),
		})

		L.CreateTable(0, 0)
		L.SetFunction(-1, "__index", func(L *lua.State) int {
			t := CheckTab(L)

			switch L.CheckString(2) {
			case "url":
				return t.GetURL(L)
			case "title":
				return t.GetTitle(L)
			case "viewport":
				return t.GetViewport(L)
			case "dialogs":
				//return t.GetDialogs(L)
			case "downloads":
				//return t.GetDownload(L)
			case "requests":
				//return t.GetRequest(L)
			case "responses":
				//return t.GetResponse(L)
			}
			return 0
		})
		L.SetMetatable(-2)
	}
	L.SetField(-2, "__index")

	L.SetGlobal("tab")
}
