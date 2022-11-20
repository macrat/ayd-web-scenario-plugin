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

	"github.com/macrat/ayd-web-scenario/internal/lua"
)

type Encodings struct {
	env *Environment
}

func (e Encodings) ToJSON(L *lua.State) int {
	v := L.ToAny(1)
	s := AsyncRun(e.env, L, func() (string, error) {
		bs, err := json.Marshal(v)
		return string(bs), err
	})
	L.PushString(s)
	return 1
}

func (e Encodings) FromJSON(L *lua.State) int {
	s := L.CheckString(1)
	var v any
	AsyncRun(e.env, L, func() (struct{}, error) {
		err := json.Unmarshal([]byte(s), &v)
		return struct{}{}, err
	})
	L.PushAny(v)
	return 1
}

func iiter(L *lua.State, index int) func() {
	index = L.AbsIndex(index)

	i := 1
	max := int(L.Len(index))

	return func() {
		if i > max {
			L.PushNil()
		} else {
			L.GetI(index, i)
			i++
		}
	}
}

func (e Encodings) ToCSV(L *lua.State) int {
	var next func()

	CSV := 1
	OPT := 2
	ROW := 3

	switch L.Type(CSV) {
	case lua.Table:
		next = iiter(L, CSV)
	case lua.Function:
		next = func() {
			L.PushNil()
			L.Copy(CSV, -1)
			if err := L.Call(0, 1); err != nil {
				L.Error(1, err)
			}
			return
		}
	default:
		L.ArgErrorf(CSV, "table or function expected, got %s", L.Type(CSV))
	}

	var b strings.Builder
	w := csv.NewWriter(&b)

	var header []string
	noheader := false

	switch L.Type(OPT) {
	case lua.Table:
		l := int(L.Len(OPT))

		for i := 1; i <= l; i++ {
			L.GetI(OPT, i)
			header = append(header, L.ToString(-1))
			L.Pop(1)
		}
		HandleError(L, w.Write(header))
	case lua.Boolean:
		noheader = !L.ToBoolean(OPT)
	case lua.Nil:
	default:
		L.ArgErrorf(OPT, "table, boolean, or nil expected, but got %s", L.Type(OPT))
	}

	for i := 0; ; i++ {
		next()

		switch L.Type(ROW) {
		case lua.Nil:
			w.Flush()
			L.PushString(b.String())
			return 1
		case lua.Table:
			if header == nil && !noheader {
				var values []string

				L.PushNil()
				for L.Next(ROW) {
					if L.Type(-2) == lua.String {
						key := L.ToString(-2)

						if key != "" {
							header = append(header, key)
							values = append(values, L.ToString(-1))
						}
					}
					L.Pop(1)
				}
				sort.Slice(values, func(i, j int) bool {
					return header[i] < header[j]
				})
				sort.Strings(header)
				HandleError(L, w.Write(header))
				HandleError(L, w.Write(values))
			} else {
				var xs []string
				if noheader {
					L.PushNil()
					for L.Next(ROW) {
						xs = append(xs, L.ToString(-1))
						L.Pop(1)
					}
				} else {
					for i, col := range header {
						L.GetField(ROW, col)

						if L.Type(-1) == lua.Nil {
							L.Pop(1)
							L.GetI(ROW, i+1)
						}

						xs = append(xs, L.ToString(-1))
						L.Pop(1)
					}
				}
				HandleError(L, w.Write(xs))
			}
			return 1
		default:
			L.Errorf(ROW, "table or nil expected, got %s", L.Type(1))
			return 0
		}

		L.Pop(1)
	}
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

func (e Encodings) FromCSV(L *lua.State) int {
	c := csv.NewReader(strings.NewReader(L.CheckString(1)))

	useHeader := true
	switch L.Type(2) {
	case lua.Boolean:
		useHeader = L.ToBoolean(2)
	case lua.Nil:
	default:
		L.ArgErrorf(2, "boolean expected, got %s", L.Type(2))
	}

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

	L.PushFunction(func(L *lua.State) int {
		e.env.Yield()

		xs, err := c.Read()
		if err == io.EOF {
			return 0
		}
		HandleError(L, err)

		if useHeader {
			L.CreateTable(0, len(header))
			for i, val := range xs {
				L.SetString(-1, header[i], val)
			}
		} else {
			L.CreateTable(len(xs), 0)
			for i, val := range xs {
				L.PushString(val)
				L.SetI(-2, i+1)
			}
		}

		return 1
	})

	if useHeader {
		L.CreateTable(0, len(header))
		for i, h := range header {
			L.PushString(h)
			L.SetI(-2, i+1)
		}

		return 2
	} else {
		return 1
	}
}

func encodeXML(L *lua.State, index int, enc *xml.Encoder) {
	L.GetI(index, 1)
	tag := L.ToString(index)
	if tag == "" {
		panic(errors.New("the first element of table should be a string."))
	}

	start := xml.StartElement{Name: xml.Name{Local: string(tag)}}
	end := xml.EndElement{Name: xml.Name{Local: string(tag)}}

	encode := func(t xml.Token) {
		if err := enc.EncodeToken(t); err != nil {
			panic(err)
		}
	}

	L.PushInteger(1)
	for L.Next(index) {
		if !L.IsInteger(-2) {
			L.PushNil()
			L.Copy(-3, -1)
			key := L.ToString(-1)
			L.Pop(1)

			start.Attr = append(start.Attr, xml.Attr{
				Name:  xml.Name{Local: key},
				Value: L.ToString(-1),
			})
		}
		L.Pop(1)
	}

	encode(start)

	l := int(L.Len(index))
	for i := 1; i <= l; i++ {
		L.GetI(index, i)
		switch L.Type(-1) {
		case lua.Nil:
		case lua.Table:
			encodeXML(L, L.GetTop(), enc)
		default:
			encode(xml.CharData(L.ToString(-1)))
		}
		L.Pop(1)
	}

	encode(end)
}

func (e Encodings) ToXML(L *lua.State) int {
	e.env.Yield()

	defer func() {
		err := recover()
		if e, ok := err.(error); ok {
			L.Error(1, e)
		} else if err != nil {
			L.Errorf(1, "%s", err)
		}
	}()

	L.AssertType(1, lua.Table)

	var b strings.Builder

	enc := xml.NewEncoder(&b)

	encodeXML(L, 1, enc)

	HandleError(L, enc.Flush())

	L.PushString(b.String())
	return 1
}

func decodeXML(L *lua.State, dec *xml.Decoder, start xml.StartElement) {
	L.CreateTable(0, 0)
	table := L.AbsIndex(-1)

	L.PushString(start.Name.Local)
	L.SetI(table, 1)

	for _, attr := range start.Attr {
		L.SetString(-1, attr.Name.Local, attr.Value)
	}

	index := 2
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return
		} else if err != nil {
			L.Error(1, err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			decodeXML(L, dec, t)
			L.SetI(table, index)
			index++
		case xml.EndElement:
			return
		case xml.CharData:
			s := strings.TrimSpace(string(t))
			if s != "" {
				L.PushString(s)
				L.SetI(table, index)
				index++
			}
		}
	}
}

func (e Encodings) FromXML(L *lua.State) int {
	dec := xml.NewDecoder(strings.NewReader(L.CheckString(1)))

	for i := 0; ; i++ {
		e.env.Yield()

		tok, err := dec.Token()
		if err == io.EOF {
			return i
		}
		HandleError(L, err)

		if start, ok := tok.(xml.StartElement); ok {
			decodeXML(L, dec, start)
		}
	}
}

func RegisterEncodings(env *Environment, L *lua.State) {
	register := func(name string, f lua.GFunction) {
		L.PushFunction(f)
		L.SetGlobal(name)
	}

	register("tojson", Encodings{env}.ToJSON)
	register("fromjson", Encodings{env}.FromJSON)

	register("tocsv", Encodings{env}.ToCSV)
	register("fromcsv", Encodings{env}.FromCSV)

	register("toxml", Encodings{env}.ToXML)
	register("fromxml", Encodings{env}.FromXML)
}
