package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/gopher-lua"
)

func TestUnpackValue(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		s    string
		want interface{}
	}{
		{`nil`, nil},
		{`true`, true},
		{`false`, false},
		{`1`, 1.0},
		{`"hello"`, "hello"},
		{`{"hello", "world"}`, []interface{}{"hello", "world"}},
		{`{hello="world", [1]="one"}`, map[string]interface{}{"hello": "world", "1": "one"}},
		{`{[1]="one"}`, []interface{}{"one"}},
		{`{[1.1]="one"}`, map[string]interface{}{"1.1": "one"}},
		{`{[2]="two"}`, map[string]interface{}{"2": "two"}},
		{`{array={3, 2, "one"}}`, map[string]interface{}{"array": []interface{}{3.0, 2.0, "one"}}},
		{`{{hello="world", foo="bar"}, true}`, []interface{}{map[string]interface{}{"hello": "world", "foo": "bar"}, true}},
	}

	for _, tt := range tests {
		AssertLuaLine(t, L, tt.s, tt.want)
	}

	fun := `function() print('hello!') end`
	if v, ok := DoLuaLine(L, fun).(string); !ok || !strings.HasPrefix(v, "function: 0x") {
		t.Errorf("%s\n%v", fun, v)
	}
}

func TestRegisterLogger(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	l := &Logger{}
	RegisterLogger(L, l)

	tests := []struct {
		s    string
		want []string
	}{
		{`print("hello")`, []string{"hello"}},
		{`print({hello="world", foo="bar"})`, []string{`{"foo":"bar","hello":"world"}`}},
		{`print(1)`, []string{"1"}},
		{`print(2, 4, 8); print(16)`, []string{"[2,4,8]", "16"}},
		{`print("hello", 123)`, []string{`["hello",123]`}},
		{`print(); print()`, []string{}},
	}

	for _, tt := range tests {
		l.Logs = []string{}

		L.DoString(tt.s)

		if diff := cmp.Diff(l.Logs, tt.want); diff != "" {
			t.Errorf("%s\n%s", tt.s, diff)
		}
	}
}
