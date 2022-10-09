package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/macrat/ayd/lib-ayd"
	"github.com/yuin/gopher-lua"
)

type Logger struct {
	Debug  bool
	Logs   []string
	Status ayd.Status
	Extra  map[string]interface{}
}

func (l *Logger) Print(values ...interface{}) {
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

func (l *Logger) AsString() string {
	return strings.Join(l.Logs, "\n")
}

func (l *Logger) AsRecord() ayd.Record {
	return ayd.Record{
		Status:  l.Status,
		Message: l.AsString(),
		Extra:   l.Extra,
	}
}

func (l *Logger) SetStatus(status string) {
	l.Status = ayd.ParseStatus(status)

	if l.Debug {
		fmt.Fprintf(os.Stdout, "::status::%s\n", l.Status)
	}
}

func (l *Logger) SetExtra(k string, v interface{}) {
	if l.Extra == nil {
		l.Extra = make(map[string]interface{})
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
			var xs []interface{}
			for i := 2; i <= L.GetTop(); i++ {
				xs = append(xs, UnpackLValue(L.Get(i)))
			}
			logger.Print(xs...)
			return 0
		},
	}))
	L.SetGlobal("print", tbl)
}

func asArray(t *lua.LTable) ([]interface{}, bool) {
	isArray := true
	values := make(map[int]lua.LValue)
	t.ForEach(func(k, v lua.LValue) {
		if n, ok := k.(lua.LNumber); ok {
			if math.Mod(float64(n), 1) != 0 {
				isArray = false
			} else {
				values[int(n)] = v
			}
		} else {
			isArray = false
		}
	})
	if !isArray {
		return nil, false
	}
	result := make([]interface{}, len(values))
	for i := 1; i <= len(values); i++ {
		v, ok := values[i]
		if !ok {
			return nil, false
		}
		result[i-1] = UnpackLValue(v)
	}
	return result, true
}

func UnpackLValue(v lua.LValue) interface{} {
	switch x := v.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return x == lua.LTrue
	case lua.LNumber:
		return float64(x)
	case lua.LString:
		return string(x)
	case *lua.LTable:
		if array, ok := asArray(x); ok {
			return array
		}

		values := make(map[string]interface{})
		x.ForEach(func(k, v lua.LValue) {
			values[k.String()] = UnpackLValue(v)
		})
		return values
	default:
		return x.String()
	}
}