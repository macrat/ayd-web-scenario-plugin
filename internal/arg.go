package webscenario

import (
	"path/filepath"
	"time"

	"github.com/macrat/ayd/lib-ayd"
	"github.com/yuin/gopher-lua"
)

type Arg struct {
	Mode      string
	Args      []string
	Target    *ayd.URL
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

func URLToTable(L *lua.LState, u *ayd.URL) *lua.LTable {
	tbl := L.NewTable()

	L.SetField(tbl, "url", lua.LString(u.String()))
	if u.User != nil {
		L.SetField(tbl, "username", lua.LString(u.User.Username()))
		if p, ok := u.User.Password(); ok {
			L.SetField(tbl, "password", L.NewFunction(func(L *lua.LState) int {
				L.Push(lua.LString(p))
				return 1
			}))
		}
	}

	qs := L.NewTable()
	L.SetField(tbl, "query", qs)
	for k, v := range u.ToURL().Query() {
		L.SetField(qs, k, lua.LString(v[len(v)-1]))
	}

	L.SetField(tbl, "fragment", lua.LString(u.Fragment))

	return tbl
}

func (a Arg) Register(L *lua.LState) {
	tbl := L.NewTable()

	for _, x := range a.Args {
		tbl.Append(lua.LString(x))
	}

	L.SetField(tbl, "target", URLToTable(L, a.Target))
	L.SetField(tbl, "debug", lua.LBool(a.Debug))
	L.SetField(tbl, "head", lua.LBool(a.Head))
	L.SetField(tbl, "recording", lua.LBool(a.Recording))

	L.SetGlobal("arg", tbl)
}
