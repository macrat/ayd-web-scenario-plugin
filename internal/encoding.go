package webscenario

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/yuin/gopher-lua"
)

type Encodings struct {
	env *Environment
}

func (e Encodings) ToJSON(L *lua.LState) int {
	v := L.Get(1)
	s := AsyncRun(e.env, L, func() (string, error) {
		bs, err := json.Marshal(UnpackLValue(v))
		return string(bs), err
	})
	L.Push(lua.LString(s))
	return 1
}

func (e Encodings) FromJSON(L *lua.LState) int {
	s := L.CheckString(1)
	var v any
	AsyncRun(e.env, L, func() (struct{}, error) {
		err := json.Unmarshal([]byte(s), &v)
		return struct{}{}, err
	})
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

func (e Encodings) ToCSV(L *lua.LState) int {
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
		e.env.Yield()

		if keep != nil {
			s, err := stringsToCSV(keep)
			HandleError(L, err)
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
				HandleError(L, err)
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
				HandleError(L, err)
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

func newReaderFromLFunction(L *lua.LState, f *lua.LFunction) *iterReader {
	return &iterReader{
		Fn: func() lua.LValue {
			L.Push(f)
			L.Call(0, 1)
			return L.Get(-1)
		},
	}
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

func checkReader(L *lua.LState, index int) io.Reader {
	var r io.Reader

	switch x := L.Get(index).(type) {
	case lua.LString:
		r = strings.NewReader(string(x))
	case *lua.LTable:
		r = &iterReader{
			Fn: iiter(x),
		}
	case *lua.LFunction:
		r = newReaderFromLFunction(L, x)
	default:
		L.ArgError(1, "string, table, or iterator function expected.")
	}

	return r
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

func (e Encodings) FromCSV(L *lua.LState) int {
	r := checkReader(L, 1)

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
		HandleError(L, err)

		header = uniqueHeader(header)
	}

	L.Push(L.NewFunction(func(L *lua.LState) int {
		e.env.Yield()

		xs, err := c.Read()
		if err == io.EOF {
			L.Push(lua.LNil)
			return 1
		}
		HandleError(L, err)

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

func encodeXML(enc *xml.Encoder, t *lua.LTable) {
	k, v := t.Next(lua.LNil)
	kn, kok := k.(lua.LNumber)
	vs, vok := v.(lua.LString)
	if !kok || kn != 1 || !vok || vs == "" {
		panic(errors.New("the first element of table should be a string."))
	}

	start := xml.StartElement{Name: xml.Name{Local: string(vs)}}
	end := xml.EndElement{Name: xml.Name{Local: string(vs)}}

	var children []lua.LValue
	for {
		k, v = t.Next(k)
		if v.Type() == lua.LTNil {
			break
		}
		if _, ok := k.(lua.LNumber); ok {
			children = append(children, v)
		} else {
			start.Attr = append(start.Attr, xml.Attr{
				Name:  xml.Name{Local: lua.LVAsString(k)},
				Value: lua.LVAsString(v),
			})
		}
	}

	encode := func(t xml.Token) {
		err := enc.EncodeToken(t)
		if err != nil {
			panic(err)
		}
	}

	encode(start)

	for _, child := range children {
		switch v := child.(type) {
		case *lua.LNilType:
		case *lua.LTable:
			encodeXML(enc, v)
		case lua.LString:
			encode(xml.CharData(v))
		default:
			encode(xml.CharData(LValueToString(v)))
		}
	}

	encode(end)
}

func (e Encodings) ToXML(L *lua.LState) int {
	defer func() {
		err := recover()
		if err != nil {
			L.RaiseError("%s", err)
		}
	}()

	v := L.CheckTable(1)

	var b strings.Builder

	AsyncRun(e.env, L, func() (struct{}, error) {
		enc := xml.NewEncoder(&b)

		encodeXML(enc, v)

		return struct{}{}, enc.Flush()
	})

	L.Push(lua.LString(b.String()))
	return 1
}

func decodeXML(dec *xml.Decoder, start xml.StartElement, L *lua.LState) *lua.LTable {
	tbl := L.NewTable()

	tbl.Append(lua.LString(start.Name.Local))

	for _, attr := range start.Attr {
		L.SetField(tbl, attr.Name.Local, lua.LString(attr.Value))
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return tbl
		} else if err != nil {
			L.RaiseError("%s", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			tbl.Append(decodeXML(dec, t, L))
		case xml.EndElement:
			return tbl
		case xml.CharData:
			s := strings.TrimSpace(string(t))
			if s != "" {
				tbl.Append(lua.LString(s))
			}
		}
	}
}

func (e Encodings) FromXML(L *lua.LState) int {
	e.env.Yield()

	r := checkReader(L, 1)
	dec := xml.NewDecoder(r)

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			L.Push(lua.LNil)
			return 1
		}
		HandleError(L, err)

		if start, ok := tok.(xml.StartElement); ok {
			tbl := decodeXML(dec, start, L)
			L.Push(tbl)
			break
		}
	}

	return 1
}

func RegisterEncodings(env *Environment) {
	env.RegisterFunction("tojson", Encodings{env}.ToJSON)
	env.RegisterFunction("fromjson", Encodings{env}.FromJSON)

	env.RegisterFunction("tocsv", Encodings{env}.ToCSV)
	env.RegisterFunction("fromcsv", Encodings{env}.FromCSV)

	env.RegisterFunction("toxml", Encodings{env}.ToXML)
	env.RegisterFunction("fromxml", Encodings{env}.FromXML)
}
