package webscenario

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/yuin/gopher-lua"
)

type Encodings Environment

func (e *Encodings) ToJSON(L *lua.LState) int {
	bs, err := json.Marshal(UnpackLValue(L.Get(1)))
	(*Environment)(e).HandleError(err)
	L.Push(lua.LString(string(bs)))
	return 1
}

func (e *Encodings) FromJSON(L *lua.LState) int {
	var v any
	json.Unmarshal([]byte(L.CheckString(1)), &v)
	L.Push(PackLValue(L, v))
	return 1
}

func ipairs(tbl *lua.LTable, f func(k, v lua.LValue)) {
	for i := 0; ; i++ {
		key := lua.LValue(lua.LNil)
		if i > 0 {
			key = lua.LNumber(i)
		}

		k, v := tbl.Next(key)
		if v.Type() == lua.LTNil {
			return
		}
		if i == 0 {
			if n, ok := k.(lua.LNumber); ok && n != 1 {
				return
			}
		}

		f(k, v)
	}
}

func iiter(tbl *lua.LTable) func() lua.LValue {
	i := 0
	finished := false
	return func() lua.LValue {
		if finished {
			return lua.LNil
		}

		key := lua.LValue(lua.LNil)
		if i > 0 {
			key = lua.LNumber(i)
		}
		i++

		_, v := tbl.Next(key)
		if v.Type() == lua.LTNil {
			finished = true
		}

		return v
	}
}

func stringsToCSV(ss []string) (string, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)
	if err := w.Write(ss); err != nil {
		return "", err
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n"), nil
}

func (e *Encodings) ToCSV(L *lua.LState) int {
	var next func() lua.LValue

	switch v := L.Get(1).(type) {
	case *lua.LTable:
		next = iiter(v)
	case *lua.LFunction:
		next = func() lua.LValue {
			L.Push(v)
			L.Call(0, 1)
			return L.Get(-1)
		}
	default:
		L.ArgError(1, "table or function expected.")
	}

	var header, keep []string
	noheader := false

	switch h := L.Get(2).(type) {
	case *lua.LTable:
		ipairs(h, func(k, v lua.LValue) {
			header = append(header, lua.LVAsString(v))
			keep = header
		})
	case lua.LBool:
		noheader = !bool(h)
	case *lua.LNilType:
	default:
		L.ArgError(2, "header property should be a nil or a table.")
	}

	L.Push(L.NewFunction(func(L *lua.LState) int {
		if keep != nil {
			s, err := stringsToCSV(keep)
			(*Environment)(e).HandleError(err)
			L.Push(lua.LString(s))
			keep = nil
			return 1
		}

		switch row := next().(type) {
		case *lua.LNilType:
			return 0
		case *lua.LTable:
			if header == nil && !noheader {
				row.ForEach(func(k, v lua.LValue) {
					x := lua.LVAsString(k)
					if x != "" {
						header = append(header, x)
						keep = append(keep, lua.LVAsString(v))
					}
				})
				sort.Slice(keep, func(i, j int) bool {
					return keep[i] < keep[j]
				})
				sort.Strings(header)
				s, err := stringsToCSV(header)
				(*Environment)(e).HandleError(err)
				L.Push(lua.LString(s))
			} else {
				var xs []string
				if noheader {
					ipairs(row, func(k, v lua.LValue) {
						xs = append(xs, lua.LVAsString(v))
					})
				} else {
					for i, col := range header {
						v := L.GetField(row, col)
						if v.Type() == lua.LTNil {
							v = row.RawGetInt(i + 1)
						}
						xs = append(xs, lua.LVAsString(v))
					}
				}
				s, err := stringsToCSV(xs)
				(*Environment)(e).HandleError(err)
				L.Push(lua.LString(s))
			}
			return 1
		default:
			L.RaiseError("a table expected.")
			return 0
		}
	}))

	return 1
}

type iterReader struct {
	done bool
	buf  []byte
	Fn   func() lua.LValue
}

func (ir *iterReader) Read(p []byte) (n int, err error) {
	if ir.done {
		return 0, io.EOF
	}

	for len(p) > len(ir.buf) {
		v := ir.Fn()
		if v.Type() == lua.LTNil {
			ir.done = true
			if ir.buf == nil {
				err = io.EOF
			} else {
				copy(p, ir.buf)
				n = len(ir.buf)
				ir.buf = nil
			}
			return
		}
		ir.buf = append(ir.buf, []byte(lua.LVAsString(v)+"\n")...)
	}
	copy(p, ir.buf)
	n = len(p)
	ir.buf = ir.buf[n:]
	return
}

func uniqueHeader(xs []string) []string {
	contains := func(x string, xs []string) bool {
		for _, y := range xs {
			if x == y {
				return true
			}
		}
		return false
	}
	rename := func(x string, xs []string) string {
		if !contains(x, xs) {
			return x
		}

		i := 1
		for {
			candidate := fmt.Sprintf("%s_%d", x, i)
			if !contains(candidate, xs) {
				return candidate
			}
			i++
		}
	}

	var rs []string
	for _, x := range xs {
		rs = append(rs, rename(x, rs))
	}
	return rs
}

func (e *Encodings) FromCSV(L *lua.LState) int {
	var r io.Reader

	switch x := L.Get(1).(type) {
	case lua.LString:
		r = strings.NewReader(string(x))
	case *lua.LTable:
		r = &iterReader{
			Fn: iiter(x),
		}
	case *lua.LFunction:
		r = &iterReader{
			Fn: func() lua.LValue {
				L.Push(x)
				L.Call(0, 1)
				return L.Get(-1)
			},
		}
	default:
		L.ArgError(1, "string, table, or iterator function expected.")
	}

	useHeader := true
	switch x := L.Get(2).(type) {
	case lua.LBool:
		useHeader = bool(x)
	case *lua.LNilType:
	default:
		L.ArgError(2, "a boolean expected.")
	}

	c := csv.NewReader(r)

	var header []string
	if useHeader {
		var err error
		header, err = c.Read()
		if err == io.EOF {
			return 0
		}
		(*Environment)(e).HandleError(err)

		header = uniqueHeader(header)
	}

	L.Push(L.NewFunction(func(L *lua.LState) int {
		xs, err := c.Read()
		if err == io.EOF {
			L.Push(lua.LNil)
			return 1
		}
		(*Environment)(e).HandleError(err)

		tbl := L.NewTable()
		if useHeader {
			for i, val := range xs {
				L.SetField(tbl, header[i], lua.LString(val))
			}
		} else {
			for _, val := range xs {
				tbl.Append(lua.LString(val))
			}
		}
		L.Push(tbl)

		return 1
	}))

	if useHeader {
		lheader := L.NewTable()
		for _, h := range header {
			lheader.Append(lua.LString(h))
		}
		L.Push(lheader)

		return 2
	} else {
		return 1
	}
}

func RegisterEncodings(env *Environment) {
	env.RegisterFunction("tojson", (*Encodings)(env).ToJSON)
	env.RegisterFunction("fromjson", (*Encodings)(env).FromJSON)

	env.RegisterFunction("tocsv", (*Encodings)(env).ToCSV)
	env.RegisterFunction("fromcsv", (*Encodings)(env).FromCSV)
}