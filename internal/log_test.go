package webscenario

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/gopher-lua"
)

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
