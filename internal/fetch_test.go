package webscenario

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/macrat/ayd-web-scenario/internal/lua"
)

func TestUnpackFetchHeader(t *testing.T) {
	tests := []struct {
		Input string
		Want  http.Header
	}{
		{
			`{["X-Hello"]="world", ["Content-Type"]="text/json"}`,
			http.Header{
				"X-Hello":      []string{"world"},
				"Content-Type": []string{"text/json"},
			},
		},
		{
			`{"nah", ["X-Foo"]={"bar", "baz"}}`,
			http.Header{
				"X-Foo": []string{"bar", "baz"},
			},
		},
		{
			`{num=123}`,
			http.Header{
				"Num": []string{"123"},
			},
		},
		{
			`{}`,
			http.Header{},
		},
		{
			`nil`,
			http.Header{},
		},
	}

	L, err := lua.NewState()
	if err != nil {
		t.Fatalf("failed to prepare lua: %s", err)
	}
	defer L.Close()

	for _, tt := range tests {
		if err := L.LoadString("return " + tt.Input); err != nil {
			t.Errorf("failed to prepare test input: %s\n%s", err, tt.Input)
			continue
		}

		err := L.Call(0, 1)
		if err != nil {
			t.Errorf("failed to call test script: %s\n%s", err, tt.Input)
			continue
		}

		actual, err := ToFetchHeader(L, -1)
		L.Pop(1)
		if err != nil {
			t.Errorf("unexpected error: %s\n%s", err, tt.Input)
			continue
		}

		if diff := cmp.Diff(tt.Want, actual); diff != "" {
			t.Errorf("unexpected header:\n%s\n%s", tt.Input, diff)
		}
	}
}
