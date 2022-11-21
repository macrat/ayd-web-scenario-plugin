package webscenario

import (
	"bufio"
	"fmt"
	"io"

	"github.com/macrat/ayd-web-scenario/internal/lua"
)

func PushFileLikeMeta(L *lua.State, r io.Reader) {
	L.CreateTable(0, 0)

	L.CreateTable(0, 1)
	{
		L.GetTypeMetatable("filelike")
		L.SetMetatable(-2)

		L.PushUserdata(bufio.NewReader(r))
		L.SetField(-2, "_reader")
	}
	L.SetField(-2, "__index")
}

func RegisterFileLike(L *lua.State) {
	checkFileReader := func(L *lua.State) *bufio.Reader {
		L.AssertType(1, lua.Table)

		if typ := L.GetField(1, "_reader"); typ != lua.Userdata {
			L.ArgErrorf(1, "_reader field is invalid, got %s", typ)
		}
		r, ok := L.ToUserdata(-1).(*bufio.Reader)
		if !ok {
			L.ArgErrorf(1, "_reader field is invalid")
		}
		L.Pop(1)
		return r
	}

	L.NewTypeMetatable("filelike")
	{
		L.CreateTable(0, 2)
		L.SetFuncs(-1, map[string]lua.GFunction{
			"lines": func(L *lua.State) int {
				r := checkFileReader(L)

				L.PushFunction(func(L *lua.State) int {
					buf, _, err := r.ReadLine()
					if err == io.EOF {
						return 0
					} else if err != nil {
						L.Error(1, err)
					}
					L.PushString(string(buf))
					return 1
				})
				return 1
			},
			"read": func(L *lua.State) int {
				r := checkFileReader(L)
				n := L.GetTop()

				if n == 1 {
					L.PushString("l")
					n = 2
				}

				for i := 2; i <= n; i++ {
					switch L.Type(i) {
					case lua.String:
						format := L.ToString(i)
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
								L.Error(1, err)
							}
							L.PushNumber(buf)
						case "a":
							buf, err := io.ReadAll(r)
							if err == io.EOF {
								return i - 2
							} else if err != nil {
								L.Error(1, err)
							}
							L.PushString(string(buf))
						case "l", "L", "":
							buf, _, err := r.ReadLine()
							if err == io.EOF {
								return i - 2
							} else if err != nil {
								L.Error(1, err)
							}
							if format[:1] == "L" {
								buf = append(buf, '\n')
							}
							L.PushString(string(buf))
						default:
							L.ArgErrorf(i, "invalid format %q", L.ToString(i)) // don't use format because it drops "*".
						}
					case lua.Number:
						buf := make([]byte, int(L.ToNumber(i)))
						_, err := r.Read(buf)
						if err == io.EOF {
							return i - 2
						} else if err != nil {
							L.Error(1, err)
						}
						L.PushString(string(buf))
					default:
						L.ArgErrorf(i, "string or number expected, got %s", L.Type(i))
					}
				}

				return n - 1
			},
		})
		L.SetField(-2, "__index")
	}
	L.Pop(1)
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
