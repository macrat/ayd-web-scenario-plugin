package webscenario

import (
	"strings"

	"github.com/chromedp/chromedp/kb"
	"github.com/macrat/ayd-web-scenario/internal/lua"
)

func RegisterKey(L *lua.State) {
	L.CreateTable(0, len(kb.Keys))

	for key, info := range kb.Keys {
		name := strings.ToLower(string(info.Code[0])) + info.Code[1:]
		L.SetString(-1, name, string(key))
	}

	L.SetGlobal("key")
}
