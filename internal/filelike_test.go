package webscenario

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/gopher-lua"
)

func TestAsFileLikeMeta(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		Script string
		Want   []lua.LValue
	}{
		{"return f:read()", []lua.LValue{lua.LString("123hello world")}},
		{"return f:read('*line')", []lua.LValue{lua.LString("123hello world")}},
		{"return f:read('line')", []lua.LValue{lua.LString("123hello world")}},
		{"return f:read('l')", []lua.LValue{lua.LString("123hello world")}},

		{"return f:read('*number')", []lua.LValue{lua.LNumber(123)}},
		{"return f:read('*n')", []lua.LValue{lua.LNumber(123)}},
		{"return f:read('number')", []lua.LValue{lua.LNumber(123)}},

		{"return f:read('*all')", []lua.LValue{lua.LString("123hello world\nfoo bar\nbaz\n")}},
		{"return f:read('all')", []lua.LValue{lua.LString("123hello world\nfoo bar\nbaz\n")}},

		{"return f:read(5)", []lua.LValue{lua.LString("123he")}},
		{"return f:read(2, 5)", []lua.LValue{lua.LString("12"), lua.LString("3hell")}},

		{"return f:read('n', 'l')", []lua.LValue{lua.LNumber(123), lua.LString("hello world")}},
		{"return f:read('L', 'l')", []lua.LValue{lua.LString("123hello world\n"), lua.LString("foo bar")}},

		{"return f:lines()()", []lua.LValue{lua.LString("123hello world")}},
		{"return f:lines()(), f:read()", []lua.LValue{lua.LString("123hello world"), lua.LString("foo bar")}},
		{"return f:lines()(), f:lines()()", []lua.LValue{lua.LString("123hello world"), lua.LString("foo bar")}},
		{"l = f:lines(); return l(), l()", []lua.LValue{lua.LString("123hello world"), lua.LString("foo bar")}},
	}

	RegisterFileLike(L)

	for _, tt := range tests {
		r := strings.NewReader("123hello world\nfoo bar\nbaz\n")

		f := L.NewTable()
		L.SetMetatable(f, AsFileLikeMeta(L, r))
		L.SetGlobal("f", f)

		err := L.DoString(tt.Script)
		if err != nil {
			t.Errorf("%s => %s", tt.Script, err)
			continue
		}

		var v []lua.LValue
		for i := 1; i <= L.GetTop(); i++ {
			v = append(v, L.Get(i))
		}
		L.Pop(L.GetTop())
		if diff := cmp.Diff(tt.Want, v); diff != "" {
			t.Errorf("%s =>\n%s", tt.Script, diff)
		}
	}
}
