package webscenario

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/macrat/ayd/lib-ayd"
	"github.com/yuin/gopher-lua"
)

type Logger struct {
	sync.Mutex

	Stream io.Writer
	Debug  bool
	Logs   []string
	Status ayd.Status
	Extra  map[string]any
}

func (l *Logger) Print(values ...lua.LValue) {
	l.Lock()
	defer l.Unlock()

	switch len(values) {
	case 0:
		l.Logs = append(l.Logs, "")
	case 1:
		if s, ok := values[0].(lua.LString); ok {
			l.Logs = append(l.Logs, string(s))
		} else {
			x, _ := json.Marshal(UnpackLValue(values[0]))
			l.Logs = append(l.Logs, string(x))
		}
	default:
		var xs []any
		for _, v := range values {
			xs = append(xs, UnpackLValue(v))
		}
		x, _ := json.Marshal(xs)
		l.Logs = append(l.Logs, string(x))
	}

	if l.Stream != nil {
		var ss []string
		for _, v := range values {
			if s, ok := v.(lua.LString); ok {
				ss = append(ss, string(s))
			} else {
				ss = append(ss, LValueToString(v))
			}
		}
		fmt.Fprintln(l.Stream, strings.Join(ss, "\t"))
	}
}

func (l *Logger) StartTask(where, name string) {
	if l.Stream != nil && l.Debug {
		fmt.Fprintln(l.Stream, where, name)
	}
}

func (l *Logger) HandleError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	l.Lock()
	defer l.Unlock()

	if l.Extra == nil {
		l.Extra = make(map[string]any)
	}

	var apierr *lua.ApiError
	if errors.As(err, &apierr) {
		err = errors.New(strings.TrimRight(apierr.Object.String(), "\n"))
		l.Extra["trace"] = apierr.StackTrace
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		l.Extra["error"] = "timeout"
		l.Status = ayd.StatusAborted
	} else if errors.Is(ctx.Err(), context.Canceled) {
		l.Extra["error"] = "interrupted"
		l.Status = ayd.StatusAborted
	} else {
		l.Extra["error"] = err.Error()
		l.Status = ayd.StatusFailure
	}

	if l.Stream != nil {
		fmt.Fprintln(l.Stream, l.Extra["error"])
		if t, ok := l.Extra["trace"]; ok {
			fmt.Fprintln(l.Stream, t)
		}
	}
}

func (l *Logger) AsRecord() ayd.Record {
	l.Lock()
	defer l.Unlock()

	return ayd.Record{
		Status:  l.Status,
		Message: strings.Join(l.Logs, "\n"),
		Extra:   l.Extra,
	}
}

func (l *Logger) SetStatus(status string) {
	l.Lock()
	defer l.Unlock()

	l.Status = ayd.ParseStatus(status)

	if l.Stream != nil {
		fmt.Fprintf(l.Stream, "::status::%s\n", l.Status)
	}
}

func (l *Logger) SetExtra(k string, v any) {
	l.Lock()
	defer l.Unlock()

	if l.Extra == nil {
		l.Extra = make(map[string]any)
	}
	l.Extra[k] = v

	if l.Stream != nil {
		if bs, err := json.Marshal(v); err == nil {
			fmt.Fprintf(l.Stream, "::%s::%s\n", k, string(bs))
		}
	}
}

func RegisterLogger(L *lua.LState, logger *Logger) {
	tbl := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"status": func(L *lua.LState) int {
			logger.SetStatus(strings.ToUpper(L.CheckString(1)))
			return 0
		},
		"extra": func(L *lua.LState) int {
			logger.SetExtra(L.CheckString(1), UnpackLValue(L.CheckAny(2)))
			return 0
		},
	})
	L.SetMetatable(tbl, L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"__call": func(L *lua.LState) int {
			var xs []lua.LValue
			for i := 2; i <= L.GetTop(); i++ {
				xs = append(xs, L.Get(i))
			}
			logger.Print(xs...)
			return 0
		},
	}))
	L.SetGlobal("print", tbl)
}
