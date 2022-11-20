package webscenario

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/macrat/ayd-web-scenario/internal/lua"
)

func TestPushFileLikeMeta(t *testing.T) {
	L, err := lua.NewState()
	if err != nil {
		t.Fatalf("failed to prepare lua: %s", err)
	}
	defer L.Close()

	tests := []struct {
		Script string
		Want   []any
	}{
		{"return f:read()", []any{"123hello world"}},
		{"return f:read('*line')", []any{"123hello world"}},
		{"return f:read('line')", []any{"123hello world"}},
		{"return f:read('l')", []any{"123hello world"}},

		{"return f:read('*number')", []any{123.0}},
		{"return f:read('*n')", []any{123.0}},
		{"return f:read('number')", []any{123.0}},

		{"return f:read('*all')", []any{"123hello world\nfoo bar\nbaz\n"}},
		{"return f:read('all')", []any{"123hello world\nfoo bar\nbaz\n"}},

		{"return f:read(5)", []any{"123he"}},
		{"return f:read(2, 5)", []any{"12", "3hell"}},

		{"return f:read('n', 'l')", []any{123.0, "hello world"}},
		{"return f:read('L', 'l')", []any{"123hello world\n", "foo bar"}},

		{"return f:lines()()", []any{"123hello world"}},
		{"return f:lines()(), f:read()", []any{"123hello world", "foo bar"}},
		{"return f:lines()(), f:lines()()", []any{"123hello world", "foo bar"}},
		{"l = f:lines(); return l(), l()", []any{"123hello world", "foo bar"}},

		{"return f:read(), f:read('a'), f:read(), f:read()", []any{"123hello world", "foo bar\nbaz\n", nil}},
		{"return f:read(), f:read('a'), f:lines()(), f:lines()()", []any{"123hello world", "foo bar\nbaz\n", nil}},
	}

	RegisterFileLike(L)

	for _, tt := range tests {
		r := strings.NewReader("123hello world\nfoo bar\nbaz\n")

		L.SetTop(0)
		L.CreateTable(0, 0)
		PushFileLikeMeta(L, r)
		L.SetMetatable(-2)
		L.SetGlobal("f")

		if err := L.LoadString(tt.Script); err != nil {
			t.Errorf("failed to load: %s\n%s", err, tt.Script)
			continue
		}
		if err := L.Call(0, lua.MultRet); err != nil {
			t.Errorf("failed to call: %s\n%s", err, tt.Script)
			continue
		}

		var v []any
		for i := 1; i <= L.GetTop(); i++ {
			v = append(v, L.ToAny(i))
		}
		L.Pop(L.GetTop())
		if diff := cmp.Diff(tt.Want, v); diff != "" {
			t.Errorf("%s =>\n%s", tt.Script, diff)
		}
	}
}
