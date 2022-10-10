package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/macrat/ayd/lib-ayd"
	"github.com/yuin/gopher-lua"
)

type Logger struct {
	sync.Mutex

	Debug  bool
	Logs   []string
	Status ayd.Status
	Extra  map[string]any
}

func (l *Logger) Print(values ...any) {
	l.Lock()
	defer l.Unlock()

	switch len(values) {
	case 0:
		return
	case 1:
		if s, ok := values[0].(string); ok {
			l.Logs = append(l.Logs, string(s))
		} else {
			x, _ := json.Marshal(values[0])
			l.Logs = append(l.Logs, string(x))
		}
	default:
		x, _ := json.Marshal(values)
		l.Logs = append(l.Logs, string(x))
	}

	if l.Debug {
		s := l.Logs[len(l.Logs)-1]
		fmt.Fprintln(os.Stdout, s)
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

	if l.Debug {
		fmt.Fprintf(os.Stdout, "::status::%s\n", l.Status)
	}
}

func (l *Logger) SetExtra(k string, v any) {
	l.Lock()
	defer l.Unlock()

	if l.Extra == nil {
		l.Extra = make(map[string]any)
	}
	l.Extra[k] = v

	if l.Debug {
		if bs, err := json.Marshal(v); err == nil {
			fmt.Fprintf(os.Stdout, "::%s::%s\n", k, string(bs))
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
			var xs []any
			for i := 2; i <= L.GetTop(); i++ {
				xs = append(xs, UnpackLValue(L.Get(i)))
			}
			logger.Print(xs...)
			return 0
		},
	}))
	L.SetGlobal("print", tbl)
}
