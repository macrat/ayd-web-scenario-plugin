package webscenario

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/macrat/ayd-web-scenario/internal/lua"
)

func TestRegisterLogger(t *testing.T) {
	L, err := lua.NewState()
	if err != nil {
		t.Fatalf("failed to prepare lua: %s", err)
	}
	defer L.Close()

	l := &Logger{}
	RegisterLogger(L, l)

	tests := []struct {
		s    string
		want []string
	}{
		{`print("hello")`, []string{"hello"}},
		{`print(1)`, []string{"1"}},
		{`print(2, 4, 8); print(16)`, []string{"2\t4\t8", "16"}},
		{`print("hello", 123)`, []string{"hello\t123"}},
		{`print(); print()`, []string{"", ""}},
	}

	for _, tt := range tests {
		l.Logs = []string{}

		L.DoString(tt.s)

		if diff := cmp.Diff(tt.want, l.Logs); diff != "" {
			t.Errorf("%s\n%s", tt.s, diff)
		}
	}
}
