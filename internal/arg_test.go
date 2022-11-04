package webscenario

import (
	"testing"

	"github.com/macrat/ayd/lib-ayd"
)

func TestArg_ArtifactDir(t *testing.T) {
	tests := []struct {
		want string
		base string
		arg  Arg
	}{
		{"path/to/script", "", Arg{Mode: "ayd", Target: &ayd.URL{Opaque: "./path/to/script.lua"}}},
		{"/tmp/script", "/tmp", Arg{Mode: "ayd", Target: &ayd.URL{Opaque: "path/to/script.lua"}}},
		{"script", "./", Arg{Mode: "ayd", Target: &ayd.URL{Opaque: "./path/to/script.lua"}}},
		{"out", "", Arg{Mode: "repl", Target: &ayd.URL{Opaque: "<stdin>"}}},
		{"somewhere/out", "./somewhere", Arg{Mode: "stdin", Target: &ayd.URL{Opaque: "<stdin>"}}},
	}

	for i, tt := range tests {
		if actual := tt.arg.ArtifactDir(tt.base); actual != tt.want {
			t.Errorf("%d: expected %q but got %q", i, tt.want, actual)
		}
	}
}
