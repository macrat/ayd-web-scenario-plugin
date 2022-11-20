package lua_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/macrat/ayd-web-scenario/internal/lua"
)

func Test_pcall(t *testing.T) {
	tests := []struct {
		F string
		R []any
	}{
		{`return "success on " .. msg`, []any{true, "success on wah"}},
		{`return "success", msg`, []any{true, "success", "wah"}},
		{`error("error on " .. msg)`, []any{false, "<string>:3: error on wah"}},
		{`native()`, []any{false, "<string>:3: native error"}},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			L := NewTestState(t)

			L.PushFunction(func(L *lua.State) int {
				L.Errorf(1, "native error")
				return 0
			})
			L.SetGlobal("native")

			err := L.LoadString(`
				function f(msg)
					` + tt.F + `
				end

				return pcall(f, "wah")
			`)
			if err != nil {
				t.Fatalf("failed to load: %s", err)
			}

			err = L.Call(0, lua.MultRet)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if L.GetTop() != len(tt.R) {
				t.Fatalf("unexpected length of stack: %d", L.GetTop())
			}

			for i, want := range tt.R {
				got := L.ToAny(i + 1)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("#%d: expected %#v but got %#v", i+1, want, got)
				}
			}
		})
	}
}

func Test_xpcall(t *testing.T) {
	tests := []struct {
		F string
		H string
		R []any
	}{
		{`return "success on " .. msg`, `return err`, []any{true, "success on wah"}},
		{`return "success", msg`, `return err`, []any{true, "success", "wah"}},
		{`error("oh no")`, `return err`, []any{false, "<string>:3: oh no"}},
		{`error("oh no")`, `return err .. "!!!"`, []any{false, "<string>:3: oh no!!!"}},
		{`error("oh no")`, `return "oops", err`, []any{false, "oops", "<string>:3: oh no"}},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			L := NewTestState(t)

			err := L.LoadString(`
				function f(msg)
					` + tt.F + `
				end

				function h(err)
					` + tt.H + `
				end

				return xpcall(f, h, "wah")
			`)
			if err != nil {
				t.Fatalf("failed to load: %s", err)
			}

			err = L.Call(0, lua.MultRet)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if L.GetTop() != len(tt.R) {
				t.Fatalf("unexpected length of stack: %d", L.GetTop())
			}

			for i, want := range tt.R {
				got := L.ToAny(i + 1)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("#%d: expected %#v but got %#v", i+1, want, got)
				}
			}
		})
	}
}
