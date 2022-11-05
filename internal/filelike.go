package webscenario

import (
	"bufio"
	"fmt"
	"io"

	"github.com/yuin/gopher-lua"
)

func AsFileLikeMeta(L *lua.LState, r io.Reader) *lua.LTable {
	idx := L.NewTable()
	L.SetMetatable(idx, L.GetTypeMetatable("filelike"))

	ud := L.NewUserData()
	ud.Value = bufio.NewReader(r)
	L.SetField(idx, "_reader", ud)

	meta := L.NewTable()
	L.SetField(meta, "__index", idx)
	return meta
}

func RegisterFileLike(L *lua.LState) {
	checkFileReader := func(L *lua.LState) *bufio.Reader {
		ud, ok := L.GetField(L.Get(1), "_reader").(*lua.LUserData)
		if !ok {
			L.ArgError(1, "_reader field expected")
		}
		r, ok := ud.Value.(*bufio.Reader)
		if !ok {
			L.ArgError(1, "_reader field expected")
		}
		return r
	}

	meta := L.NewTypeMetatable("filelike")
	L.SetField(meta, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"lines": func(L *lua.LState) int {
			r := checkFileReader(L)

			L.Push(L.NewFunction(func(L *lua.LState) int {
				buf, _, err := r.ReadLine()
				if err == io.EOF {
					return 0
				} else if err != nil {
					L.RaiseError("%s", err)
				}
				L.Push(lua.LString(buf))
				return 1
			}))
			return 1
		},
		"read": func(L *lua.LState) int {
			r := checkFileReader(L)
			n := L.GetTop()

			if n == 1 {
				L.Push(lua.LString("l"))
				n = 2
			}

			for i := 2; i <= n; i++ {
				switch format := L.Get(i).(type) {
				case lua.LString:
					if format[:1] == "*" {
						format = format[1:]
					}
					switch format[:1] {
					case "n":
						var buf float64
						_, err := fmt.Fscanf(r, "%f", &buf)
						if err == io.EOF {
							return i - 2
						} else if err != nil {
							L.RaiseError("%s", err)
						}
						L.Push(lua.LNumber(buf))
					case "a":
						buf, err := io.ReadAll(r)
						if err == io.EOF {
							return i - 2
						} else if err != nil {
							L.RaiseError("%s", err)
						}
						L.Push(lua.LString(buf))
					case "l", "L", "":
						buf, _, err := r.ReadLine()
						if err == io.EOF {
							return i - 2
						} else if err != nil {
							L.RaiseError("%s", err)
						}
						if format[:1] == "L" {
							buf = append(buf, '\n')
						}
						L.Push(lua.LString(buf))
					default:
						L.ArgError(i, fmt.Sprintf("invalid format %q", L.ToString(i)))
					}
				case lua.LNumber:
					buf := make([]byte, int(format))
					_, err := r.Read(buf)
					if err == io.EOF {
						return i - 2
					} else if err != nil {
						L.RaiseError("%s", err)
					}
					L.Push(lua.LString(buf))
				default:
					L.ArgError(i, "format string or number expected")
				}
			}

			return n - 1
		},
	}))
}

type DelayedReader struct {
	open   func() io.Reader
	reader io.Reader
}

func NewDelayedReader(f func() io.Reader) *DelayedReader {
	return &DelayedReader{
		open: f,
	}
}

func (r *DelayedReader) Read(p []byte) (int, error) {
	if r.reader == nil {
		r.reader = r.open()
	}
	return r.reader.Read(p)
}
