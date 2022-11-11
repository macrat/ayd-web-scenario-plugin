package webscenario

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/yuin/gopher-lua"
)

func UnpackFetchHeader(L *lua.LState, lv lua.LValue) (http.Header, error) {
	if lv.Type() == lua.LTNil {
		return http.Header{}, nil
	}

	t, ok := lv.(*lua.LTable)
	if !ok {
		return http.Header{}, errors.New("header field is expected be a table.")
	}

	h := http.Header{}

	t.ForEach(func(k, v lua.LValue) {
		key, ok := k.(lua.LString)
		if !ok {
			return
		}

		switch v := v.(type) {
		case *lua.LTable:
			ipairs(v, func(_, v lua.LValue) {
				h.Add(string(key), string(L.ToStringMeta(v).(lua.LString)))
			})
		case lua.LString:
			h.Set(string(key), string(v))
		default:
			h.Set(string(key), string(L.ToStringMeta(v).(lua.LString)))
		}
	})

	return h, nil
}

func PackFetchHeader(L *lua.LState, h http.Header) lua.LValue {
	tbl := L.NewTable()

	for k, vs := range h {
		vt := L.NewTable()
		for _, v := range vs {
			vt.Append(lua.LString(v))
		}
		L.SetField(tbl, k, vt)
	}

	return tbl
}

func PackFetchResponse(env *Environment, L *lua.LState, resp *http.Response, body io.Reader) lua.LValue {
	tbl := L.NewTable()
	L.SetMetatable(tbl, AsFileLikeMeta(L, body))

	L.SetField(tbl, "url", lua.LString(resp.Request.URL.String()))
	L.SetField(tbl, "status", lua.LNumber(resp.StatusCode))
	L.SetField(tbl, "headers", PackFetchHeader(L, resp.Header))
	L.SetField(tbl, "length", lua.LNumber(resp.ContentLength))

	return tbl
}

type CookieJar struct {
	id   int
	jar  *cookiejar.Jar
	urls map[string]*url.URL
}

func NewCookieJar(id int) (*CookieJar, error) {
	jar, err := cookiejar.New(nil)
	return &CookieJar{
		id:   id,
		jar:  jar,
		urls: make(map[string]*url.URL),
	}, err
}

func CheckCookieJar(L *lua.LState, n int) *CookieJar {
	ud := L.ToUserData(n)
	if ud == nil {
		L.ArgError(n, "cookiejar expected.")
	}

	j, ok := ud.Value.(*CookieJar)
	if !ok {
		L.ArgError(n, "cookiejar expected.")
	}

	return j
}

func (j *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.urls[u.String()] = u
	j.jar.SetCookies(u, cookies)
}

func (j *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.jar.Cookies(u)
}

func (j *CookieJar) CookiesAsLua(L *lua.LState, u *url.URL) (*lua.LTable, bool) {
	cs := j.Cookies(u)
	if len(cs) == 0 {
		return nil, false
	}

	tbl := L.NewTable()
	for _, c := range cs {
		l := L.NewTable()
		L.SetField(l, "name", lua.LString(c.Name))
		L.SetField(l, "value", lua.LString(c.Value))
		L.SetField(l, "path", lua.LString(c.Path))
		L.SetField(l, "domain", lua.LString(c.Domain))
		if !c.Expires.IsZero() {
			L.SetField(l, "expires", lua.LNumber(c.Expires.UnixMilli()))
		}
		L.SetField(l, "secure", lua.LBool(c.Secure))
		L.SetField(l, "httponly", lua.LBool(c.HttpOnly))

		var samesite string
		switch c.SameSite {
		case http.SameSiteDefaultMode:
			samesite = "default"
		case http.SameSiteLaxMode:
			samesite = "lax"
		case http.SameSiteStrictMode:
			samesite = "strict"
		case http.SameSiteNoneMode:
			samesite = "none"
		}
		L.SetField(l, "samesite", lua.LString(samesite))

		tbl.Append(l)
	}
	return tbl, true
}

func (j *CookieJar) ToLua(L *lua.LState) lua.LValue {
	v := L.NewUserData()
	v.Value = j
	v.Metatable = L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"__tostring": func(L *lua.LState) int {
			L.Push(lua.LString(fmt.Sprintf("cookiejar#%d", j.id)))
			return 1
		},
	})
	L.SetField(v.Metatable, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"all": func(L *lua.LState) int {
			j := CheckCookieJar(L, 1)

			tbl := L.NewTable()
			for s, u := range j.urls {
				if cs, ok := j.CookiesAsLua(L, u); ok {
					L.SetField(tbl, s, cs)
				}
			}
			L.Push(tbl)

			return 1
		},
		"get": func(L *lua.LState) int {
			j := CheckCookieJar(L, 1)

			u, err := url.Parse(L.ToString(2))
			if err != nil {
				L.ArgError(2, "valid url expected.")
			}

			cs, _ := j.CookiesAsLua(L, u)
			L.Push(cs)

			return 1
		},
	}))
	return v
}

func RegisterFetch(ctx context.Context, env *Environment) {
	jarID := 1

	env.RegisterFunction("fetch", func(L *lua.LState) int {
		url := L.CheckString(1)
		opts := L.OptTable(2, L.NewTable())

		header, err := UnpackFetchHeader(L, L.GetField(opts, "headers"))
		if err != nil {
			L.ArgError(2, err.Error())
		}

		var body io.Reader
		switch b := L.GetField(opts, "body").(type) {
		case *lua.LNilType:
		case lua.LString:
			body = strings.NewReader(string(b))
		case lua.LNumber:
			body = strings.NewReader(string(b.String()))
		case *lua.LFunction:
			body = newReaderFromLFunction(L, b)
		default:
			L.ArgError(2, "body field expected be a string.")
		}

		method := ""
		switch m := L.GetField(opts, "method").(type) {
		case *lua.LNilType:
		case lua.LString:
			method = string(m)
		default:
			L.ArgError(2, "method field expected be a string.")
		}
		if method == "" {
			if body != nil {
				method = "POST"
			} else {
				method = "GET"
			}
		}

		timeout := time.Duration(5 * time.Minute)
		switch t := L.GetField(opts, "timeout").(type) {
		case *lua.LNilType:
		case lua.LNumber:
			timeout = time.Duration(float64(t) * float64(time.Millisecond))
		default:
			L.ArgError(2, "timeout field expected be a string.")
		}

		var cookiejar *CookieJar
		switch s := L.GetField(opts, "cookiejar").(type) {
		case *lua.LNilType:
			var err error
			cookiejar, err = NewCookieJar(jarID)
			if err != nil {
				L.RaiseError("failed to prepare session: %s", err)
			}
			jarID++
		case *lua.LUserData:
			if j, ok := s.Value.(*CookieJar); ok {
				cookiejar = j
				break
			}
			L.ArgError(2, "session field expected session value.")
		default:
			L.ArgError(2, "session field expected session value.")
		}

		type Ret struct {
			Resp *http.Response
			Body []byte
		}
		ret := AsyncRun(env, L, func() (Ret, error) {
			c := ctx
			if timeout > 0 {
				var cancel context.CancelFunc
				c, cancel = context.WithTimeout(c, timeout)
				defer cancel()
			}

			req, err := http.NewRequestWithContext(c, method, url, body)
			if err != nil {
				return Ret{nil, nil}, err
			}
			req.Header = header

			resp, err := (&http.Client{Jar: cookiejar}).Do(req)
			if err != nil {
				return Ret{nil, nil}, err
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()

			return Ret{resp, body}, err
		})

		L.Push(PackFetchResponse(env, L, ret.Resp, bytes.NewReader(ret.Body)))
		L.Push(cookiejar.ToLua(L))

		return 2
	})
}
