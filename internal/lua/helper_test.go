package lua_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/macrat/ayd-web-scenario/internal/lua"
)

func TestState_Swap(t *testing.T) {
	L := NewTestState(t)

	get := func() string {
		var ss []string
		for i := 1; i <= L.GetTop(); i++ {
			ss = append(ss, L.ToString(i))
		}
		return strings.Join(ss, ":")
	}

	L.PushString("A")
	L.PushString("B")

	if s := get(); s != "A:B" {
		t.Errorf("unexpected order: %s", s)
	}

	L.Swap(1, 2)
	if s := get(); s != "B:A" {
		t.Errorf("unexpected order: %s", s)
	}

	L.PushString("C")

	L.Swap(3, 1)
	if s := get(); s != "C:A:B" {
		t.Errorf("unexpected order: %s", s)
	}
}

func TestState_ToAny(t *testing.T) {
	L := NewTestState(t)

	tests := []struct {
		s    string
		want any
	}{
		{`nil`, nil},
		{`true`, true},
		{`false`, false},
		{`1`, int64(1)},
		{`1.0`, float64(1.0)},
		{`"hello"`, "hello"},
		{`{"hello", "world"}`, []any{"hello", "world"}},
		{`{hello="world", [1]="one"}`, map[string]any{"hello": "world", "1": "one"}},
		{`{[1]="one"}`, []any{"one"}},
		{`{[1.1]="one"}`, map[string]any{"1.1": "one"}},
		{`{[2]="two"}`, map[string]any{"2": "two"}},
		{`{nil, "two"}`, []any{nil, "two"}},
		{`{array={3, 2.0, "one"}}`, map[string]any{"array": []any{int64(3), float64(2.0), "one"}}},
		{`{{hello="world", foo="bar"}, true}`, []any{map[string]any{"hello": "world", "foo": "bar"}, true}},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			L.SetTop(0)
			if err := L.LoadString("return " + tt.s); err != nil {
				t.Fatalf("failed to load script: %s", err)
			}

			if err := L.Call(0, 1); err != nil {
				t.Fatalf("failed to execute script: %s", err)
			}

			if diff := cmp.Diff(tt.want, L.ToAny(1)); diff != "" {
				t.Fatalf("unexpected result:\n%s", diff)
			}
		})
	}
}

func TestState_PushAny(t *testing.T) {
	tests := []struct {
		v    any
		want any
	}{
		{nil, nil},
		{[]string(nil), nil},
		{map[int]float32(nil), nil},
		{true, true},
		{false, false},
		{1, int64(1)},
		{uint64(2), int64(2)},
		{1.2, float64(1.2)},
		{1.0, float64(1.0)},
		{"hello", "hello"},
		{[]string{"hello", "world"}, []any{"hello", "world"}},
		{map[string]string{"hello": "world"}, map[string]any{"hello": "world"}},
		{map[string]string{"1": "one"}, map[string]any{"1": "one"}},
		{map[float32]string{1.1: "one"}, map[string]any{"1.1": "one"}},
		{map[int]string{1: "one"}, map[string]any{"1": "one"}},
		{map[string][]string{"array": {"hello", "world"}}, map[string]any{"array": []any{"hello", "world"}}},
		{[]map[string]float64{{"foo": 1.23}, {"bar": 4.56}}, []any{map[string]any{"foo": 1.23}, map[string]any{"bar": 4.56}}},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			L := NewTestState(t)

			L.PushAny(tt.v)
			got := L.ToAny(1)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("unexpected value\n%s", diff)
			}
		})
	}
}

func TestState_PushAny_gFunction(t *testing.T) {
	L := NewTestState(t)

	L.PushAny(func(L *lua.State) int {
		L.PushString("hello")
		return 1
	})
	if typ := L.Type(1); typ != lua.Function {
		t.Fatalf("expected function but got %s", typ)
	}
	if err := L.Call(0, 1); err != nil {
		t.Fatalf("failed to call: %s", err)
	}
	if s := L.ToString(1); s != "hello" {
		t.Fatalf("unexpected result: %q", s)
	}
}

func TestState_PushAny_normalFunction(t *testing.T) {
	L := NewTestState(t)

	L.PushAny(func() string {
		return "hello"
	})
	if typ := L.Type(1); typ != lua.Userdata {
		t.Fatalf("expected userdata but got %s", typ)
	}

	f, ok := L.ToAny(1).(func() string)
	if !ok {
		t.Fatalf("failed to get pushed function")
	}

	if s := f(); s != "hello" {
		t.Fatalf("unexpected result: %q", s)
	}
}

func TestState_PushAny_userdata(t *testing.T) {
	L := NewTestState(t)

	type Cons struct {
		A int
		B int
	}

	L.PushAny(Cons{1, 7})

	if typ := L.Type(1); typ != lua.Userdata {
		t.Fatalf("expected userdata but got %s", typ)
	}

	x, ok := L.ToAny(1).(Cons)
	if !ok {
		t.Fatalf("failed to get pushed struct")
	}

	if diff := cmp.Diff(Cons{1, 7}, x); diff != "" {
		t.Fatalf("unexpected result:\n%s", diff)
	}
}

func TestState_CallWithContext(t *testing.T) {
	L := NewTestState(t)

	L.PushInteger(0)
	L.SetGlobal("count")

	err := L.LoadString(`
		while true do
			count = count + 1
		end
	`)
	if err != nil {
		t.Fatalf("failed to load test script: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = L.CallWithContext(ctx, 0, 0)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected stop by timeout but got: %s", err)
	}

	L.GetGlobal("count")
	if n := L.ToInteger(1); n == 0 {
		t.Fatalf("expected incremented count but got: %d", n)
	}
}

func TestState_Errorf(t *testing.T) {
	L := NewTestState(t)

	L.PushFunction(func(L *lua.State) int {
		L.Errorf(1, "foo %s", "bar")
		return 0
	})
	L.SetGlobal("f")

	err := L.DoString(`f()`)
	e, ok := err.(lua.Error)
	if !ok {
		t.Fatalf("unexpected error found: %s", err)
	}

	if s := e.Err.Error(); s != "foo bar" {
		t.Errorf("unexpected error message: %q", s)
	}

	if e.ChunkName != "<string>" {
		t.Errorf("unexpected chunk name: %q", e.ChunkName)
	}

	if e.CurrentLine != 1 {
		t.Errorf("unexpected current line: %d", e.CurrentLine)
	}

	trace := strings.Join([]string{
		`stack traceback:`,
		`	<string>:1: in main chunk`,
	}, "\n")
	if e.Traceback != trace {
		t.Errorf("unexpected traceback:\n%s", e.Traceback)
	}

	msg := "<string>:1: foo bar\n" + trace
	if e.Error() != msg {
		t.Errorf("unexpected message:\n%s", e.Error())
	}
}
