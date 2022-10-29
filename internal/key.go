package webscenario

import (
	"strings"

	"github.com/chromedp/chromedp/kb"
	"github.com/yuin/gopher-lua"
)

func RegisterKey(L *lua.LState) {
	tbl := L.NewTable()

	for key, info := range kb.Keys {
		if !info.Print {
			code := strings.ToLower(string(info.Code[0])) + info.Code[1:]
			L.SetField(tbl, code, lua.LString(string(key)))
		}
	}

	L.SetGlobal("key", tbl)
}
