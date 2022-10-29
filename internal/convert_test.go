package webscenario

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/gopher-lua"
)

func DoLuaLine(L *lua.LState, script string) any {
	L.DoString("return " + script)
	v := UnpackLValue(L.Get(1))
	L.Pop(1)
	return v
}

func AssertLuaLine(t *testing.T, L *lua.LState, script string, want any) {
	t.Helper()

	if diff := cmp.Diff(DoLuaLine(L, script), want); diff != "" {
		t.Errorf("%s\n%s", script, diff)
	}
}

func TestUnpackValue(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		s    string
		want any
	}{
		{`nil`, nil},
		{`true`, true},
		{`false`, false},
		{`1`, 1.0},
		{`"hello"`, "hello"},
		{`{"hello", "world"}`, []any{"hello", "world"}},
		{`{hello="world", [1]="one"}`, map[string]any{"hello": "world", "1": "one"}},
		{`{[1]="one"}`, []any{"one"}},
		{`{[1.1]="one"}`, map[string]any{"1.1": "one"}},
		{`{[2]="two"}`, map[string]any{"2": "two"}},
		{`{array={3, 2, "one"}}`, map[string]any{"array": []any{3.0, 2.0, "one"}}},
		{`{{hello="world", foo="bar"}, true}`, []any{map[string]any{"hello": "world", "foo": "bar"}, true}},
	}

	for _, tt := range tests {
		AssertLuaLine(t, L, tt.s, tt.want)
	}

	fun := `function() print('hello!') end`
	if v, ok := DoLuaLine(L, fun).(string); !ok || !strings.HasPrefix(v, "function: 0x") {
		t.Errorf("%s\n%v", fun, v)
	}
}

func TestPackValue(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		v    any
		want any
	}{
		{nil, nil},
		{[]string(nil), nil},
		{map[int]float32(nil), nil},
		{true, true},
		{false, false},
		{1, 1.0},
		{uint64(2), 2.0},
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
		got := UnpackLValue(PackLValue(L, tt.v))
		if diff := cmp.Diff(got, tt.want); diff != "" {
			t.Errorf("%d: unexpected value\nexpected: %#v\n but got: %#v", i, tt.want, got)
		}
	}

	fun := `function() print('hello!') end`
	if v, ok := DoLuaLine(L, fun).(string); !ok || !strings.HasPrefix(v, "function: 0x") {
		t.Errorf("%s\n%v", fun, v)
	}
}

func TestLValueToString(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		s    string
		want string
	}{
		{`nil`, "nil"},
		{`true`, "true"},
		{`false`, "false"},
		{`1`, "1"},
		{`"hello"`, `"hello"`},
		{`{"hello", "world"}`, `{"hello", "world"}`},
		{`{hello="world", [1]="one"}`, `{"one", hello="world"}`},
		{`{[1]="one"}`, `{"one"}`},
		{`{[1.1]="one"}`, `{[1.1]="one"}`},
		{`{[2]="two"}`, `{[2]="two"}`},
		{`{array={3, 2, "one"}}`, `{array={3, 2, "one"}}`},
	}

	for _, tt := range tests {
		L.DoString("return " + tt.s)
		v := LValueToString(L.Get(1))
		L.Pop(1)

		if v != tt.want {
			t.Errorf("unexpected result\nsource: %s\nexpected: %s\n but got: %s", tt.s, tt.want, v)
		}
	}
}
