package webscenario

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/macrat/ayd/lib-ayd"
	"github.com/yuin/gopher-lua"
)

type Logger struct {
	sync.Mutex

	Stream     io.Writer
	Debug      bool
	Logs       []string
	Status     ayd.Status
	Latency    time.Duration
	LatencySet bool
	Extra      map[string]any
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

	msg := err.Error()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		l.Status = ayd.StatusUnknown
	} else if errors.Is(ctx.Err(), context.Canceled) {
		l.Status = ayd.StatusAborted
	} else {
		l.Status = ayd.StatusFailure
	}

	l.Logs = append(l.Logs, strings.TrimRight(msg, "\n"))

	if l.Stream != nil {
		fmt.Fprint(l.Stream, msg)
	}
}

func (l *Logger) AsRecord(timestamp time.Time, latency time.Duration) ayd.Record {
	l.Lock()
	defer l.Unlock()

	r := ayd.Record{
		Time:    timestamp,
		Status:  l.Status,
		Message: strings.Join(l.Logs, "\n"),
		Latency: latency,
		Extra:   l.Extra,
	}
	if l.LatencySet {
		r.Latency = l.Latency
	}
	return r
}

func (l *Logger) SetStatus(status string) {
	l.Lock()
	defer l.Unlock()

	l.Status = ayd.ParseStatus(status)

	if l.Stream != nil {
		fmt.Fprintf(l.Stream, "::status::%s\n", l.Status)
	}
}

func (l *Logger) SetLatency(milliseconds float64) {
	l.Lock()
	defer l.Unlock()

	if milliseconds < 0 {
		milliseconds = 0
	}

	l.Latency = time.Duration(milliseconds * float64(time.Millisecond))
	l.LatencySet = true

	if l.Stream != nil {
		fmt.Fprintf(l.Stream, "::latency::%f\n", milliseconds)
	}
}

func (l *Logger) UnsetLatency() {
	l.Lock()
	defer l.Unlock()

	l.LatencySet = false

	if l.Stream != nil {
		fmt.Fprintf(l.Stream, "::latency::\n")
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
		"latency": func(L *lua.LState) int {
			if L.Get(1).Type() == lua.LTNil {
				logger.UnsetLatency()
			} else {
				logger.SetLatency(float64(L.CheckNumber(1)))
			}
			return 0
		},
		"extra": func(L *lua.LState) int {
			key := L.CheckString(1)
			switch strings.ToLower(key) {
			case "message":
				L.RaiseError("print.extra() can not set message. please use print().")
			case "status":
				L.RaiseError("print.extra() can not set status. please use print.status().")
			case "latency":
				L.RaiseError("print.extra() can not set latency. please use print.latency().")
			case "time", "target":
				L.RaiseError("print.extra() can not set %s.", key)
			default:
				value := UnpackLValue(L.CheckAny(2))
				logger.SetExtra(key, value)
			}
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
