package webscenario

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/gopher-lua"
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

	L := lua.NewState()
	defer L.Close()

	for _, tt := range tests {
		if err := L.DoString("return " + tt.Input); err != nil {
			t.Errorf("failed to prepare test input: %s\n%s", err, tt.Input)
			continue
		}

		v := L.Get(1)
		L.Pop(1)

		actual, err := UnpackFetchHeader(L, v)
		if err != nil {
			t.Errorf("unexpected error: %s\n%s", err, tt.Input)
			continue
		}

		if diff := cmp.Diff(tt.Want, actual); diff != "" {
			t.Errorf("unexpected header:\n%s\n%s", tt.Input, diff)
		}
	}
}
