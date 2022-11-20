package webscenario

import (
	"path/filepath"
	"time"

	"github.com/macrat/ayd-web-scenario/internal/lua"
	"github.com/macrat/ayd/lib-ayd"
)

type Arg struct {
	Mode      string
	Args      []string
	Target    *ayd.URL
	Alert     ayd.Record
	Timeout   time.Duration
	Debug     bool
	Head      bool
	Recording bool
}

func (a Arg) ArtifactDir(basedir string) string {
	if a.Mode == "repl" || a.Mode == "stdin" {
		return filepath.Join(basedir, "out")
	}

	path := a.Target.Path
	if a.Target.Opaque != "" {
		path = a.Target.Opaque
	}

	if basedir == "" {
		return filepath.Clean(path[:len(path)-len(filepath.Ext(path))])
	} else {
		name := filepath.Base(path[:len(path)-len(filepath.Ext(path))])
		return filepath.Join(basedir, name)
	}
}

func (a Arg) Path() string {
	if a.Mode == "repl" || a.Mode == "stdin" {
		return "<stdin>"
	} else if a.Target.Opaque != "" {
		return a.Target.Opaque
	} else {
		return a.Target.Path
	}
}

func PushURLTable(L *lua.State, u *ayd.URL) {
	L.CreateTable(0, 5)

	L.SetString(-1, "url", u.String())

	if u.User != nil {
		L.SetString(-1, "username", u.User.Username())
		if p, ok := u.User.Password(); ok {
			L.SetFunction(-1, "password", func(L *lua.State) int {
				L.PushString(p)
				return 1
			})
		}
	}

	qs := u.ToURL().Query()
	L.CreateTable(0, len(qs))
	for k, v := range u.ToURL().Query() {
		L.SetString(-1, k, v[len(v)-1])
	}
	L.SetField(-2, "query")

	L.SetString(-1, "fragment", u.Fragment)
}

func (a Arg) Register(L *lua.State) {
	L.CreateTable(len(a.Args), 0)
	TABLE := L.GetTop()

	for i, x := range a.Args {
		L.PushString(x)
		L.SetI(TABLE, i+1)
	}

	L.SetString(TABLE, "mode", a.Mode)
	PushURLTable(L, a.Target)
	L.SetField(TABLE, "target")
	L.SetBoolean(TABLE, "debug", a.Debug)
	L.SetBoolean(TABLE, "head", a.Head)
	L.SetBoolean(TABLE, "recording", a.Recording)

	if a.Alert.Target != nil {
		L.CreateTable(0, 6)
		{
			L.SetInteger(-1, "time", a.Alert.Time.UnixMilli())
			L.SetString(-1, "status", a.Alert.Status.String())
			L.SetNumber(-1, "latency", float64(a.Alert.Latency.Microseconds())/1000.0)
			L.SetString(-1, "target", a.Alert.Target.String())
			L.SetString(-1, "message", a.Alert.Message)

			L.PushAny(a.Alert.Extra)
			L.SetField(-2, "extra")
		}
		L.SetField(TABLE, "alert")
	}

	L.SetGlobal("arg")
}
