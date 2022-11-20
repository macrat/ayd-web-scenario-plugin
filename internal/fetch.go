package webscenario

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/macrat/ayd-web-scenario/internal/lua"
)

func ToFetchHeader(L *lua.State, index int) (http.Header, error) {
	TABLE := L.AbsIndex(index)
	KEY := TABLE + 1
	VALUE := TABLE + 2

	switch L.Type(TABLE) {
	case lua.Table:
	case lua.Nil:
		return http.Header{}, nil
	default:
		return http.Header{}, fmt.Errorf("header field is expected table, got %s", L.Type(-1))
	}

	h := http.Header{}

	L.PushNil()
	for L.Next(TABLE) {
		if L.Type(KEY) == lua.String {
			key := L.ToString(KEY)

			switch L.Type(VALUE) {
			case lua.Table:
				l := int(L.Len(VALUE))
				for i := 1; i <= l; i++ {
					L.GetI(VALUE, i)
					h.Add(key, L.ToString(-1))
					L.Pop(1)
				}
			default:
				h.Set(key, L.ToString(VALUE))
			}
		}
		L.Pop(1)
	}

	return h, nil
}

func PushFetchHeader(L *lua.State, h http.Header) {
	L.CreateTable(0, len(h))
	TABLE := L.GetTop()

	for k, vs := range h {
		L.CreateTable(len(vs), 0)
		for i, v := range vs {
			L.PushString(v)
			L.SetI(-2, i+1)
		}
		L.SetField(TABLE, k)
	}
}

func PushFetchResponse(env *Environment, L *lua.State, resp *http.Response, body io.Reader) {
	L.CreateTable(0, 4)

	PushFileLikeMeta(L, body)
	L.SetMetatable(-2)

	L.SetString(-1, "url", resp.Request.URL.String())
	L.SetInteger(-1, "status", int64(resp.StatusCode))
	PushFetchHeader(L, resp.Header)
	L.SetField(-2, "headers")
	L.SetInteger(-1, "length", resp.ContentLength)
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

func CheckCookieJar(L *lua.State, n int) *CookieJar {
	ud := L.ToUserdata(n)
	if ud == nil {
		L.ArgErrorf(n, "cookiejar expected, got %s", L.Type(n))
	}

	j, ok := ud.(*CookieJar)
	if !ok {
		L.ArgErrorf(n, "cookiejar expected, got %s", L.Type(n))
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

func (j *CookieJar) PushCookies(L *lua.State, u *url.URL) (ok bool) {
	cs := j.Cookies(u)
	if len(cs) == 0 {
		return false
	}

	L.CreateTable(len(cs), 0)

	for i, c := range cs {
		L.CreateTable(9, 0)

		L.SetString(-1, "name", c.Name)
		L.SetString(-1, "value", c.Value)
		L.SetString(-1, "path", c.Path)
		L.SetString(-1, "domain", c.Domain)
		if !c.Expires.IsZero() {
			L.SetInteger(-1, "expires", c.Expires.UnixMilli())
		}
		L.SetBoolean(-1, "secure", c.Secure)
		L.SetBoolean(-1, "httponly", c.HttpOnly)

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
		L.SetString(-1, "samesite", samesite)

		L.SetI(-2, i+1)
	}

	return true
}

func (j *CookieJar) PushTo(L *lua.State) {
	L.PushUserdata(j)

	L.CreateTable(0, 1)
	L.PushString(fmt.Sprintf("cookiejar#%d", j.id))
	L.SetField(-2, "__name")

	L.CreateTable(0, 2)
	L.SetFuncs(-1, map[string]lua.GFunction{
		"all": func(L *lua.State) int {
			j := CheckCookieJar(L, 1)

			L.CreateTable(0, len(j.urls))
			for s, u := range j.urls {
				if ok := j.PushCookies(L, u); ok {
					L.SetField(-2, s)
				}
			}

			return 1
		},
		"get": func(L *lua.State) int {
			j := CheckCookieJar(L, 1)

			u, err := url.Parse(L.ToString(2))
			if err != nil {
				L.ArgErrorf(2, "valid url expected, got %q (%s)", L.ToString(2), L.Type(2))
			}

			if ok := j.PushCookies(L, u); ok {
				return 1
			} else {
				return 0
			}
		},
	})
	L.SetField(-2, "__index")

	L.SetMetatable(-2)
}

func RegisterFetch(ctx context.Context, env *Environment, L *lua.State) {
	jarID := 1

	L.PushFunction(func(L *lua.State) int {
		u := L.CheckString(1)

		var header http.Header
		var body io.Reader
		var method string
		timeout := time.Duration(5 * time.Minute)
		var cookiejar *CookieJar

		if optType := L.Type(2); optType != lua.Nil {
			if optType != lua.Table {
				L.ArgErrorf(2, "table or nil expected, got %s", optType)
			}

			var err error
			L.GetField(2, "headers")
			header, err = ToFetchHeader(L, -1)
			if err != nil {
				L.ArgErrorf(2, "%s", err)
			}
			L.Pop(1)

			if L.GetField(2, "body") != lua.Nil {
				body = strings.NewReader(L.ToString(3))
			}
			L.Pop(1)

			if L.GetField(2, "method") != lua.Nil {
				method = L.ToString(3)
			}
			L.Pop(1)

			if typ := L.GetField(2, "timeout"); typ == lua.Number {
				timeout = time.Duration(L.ToNumber(3) * float64(time.Millisecond))
			} else {
				L.ArgErrorf(2, "timeout field expected string, got %s", typ)
			}
			L.Pop(1)

			switch L.GetField(2, "cookiejar") {
			case lua.Nil:
				var err error
				cookiejar, err = NewCookieJar(jarID)
				if err != nil {
					L.Errorf(1, "failed to prepare cookiejar: %w", err)
				}
				jarID++
			case lua.Userdata:
				if j, ok := L.ToUserdata(3).(*CookieJar); ok {
					cookiejar = j
					break
				}
				L.ArgErrorf(2, "cookiejar field is invalid")
			default:
				L.ArgErrorf(2, "cookiejar field is invalid, got %s", L.Type(3))
			}
			L.Pop(1)
		}

		switch method {
		case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace:
		case "":
			if body != nil {
				method = http.MethodPost
			} else {
				method = http.MethodGet
			}
		default:
			L.ArgErrorf(2, "method field is invalid, got %q", method)
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

			req, err := http.NewRequestWithContext(c, method, u, body)
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

		PushFetchResponse(env, L, ret.Resp, bytes.NewReader(ret.Body))
		cookiejar.PushTo(L)
		return 2
	})
	L.SetGlobal("fetch")
}
