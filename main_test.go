package main

import (
	"runtime"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	type Test struct {
		input string
		want  string
	}

	tests := []Test{
		{"web-scenario:/path/to/script.lua", "web-scenario:/path/to/script.lua"},
		{"web-scenario:./path/to/script.lua", "web-scenario:./path/to/script.lua"},
		{"web-scenario:path/to/script.lua", "web-scenario:path/to/script.lua"},
		{"web-scenario:///path/to/script.lua", "web-scenario:/path/to/script.lua"},
		{"web-scenario://localhost/path/to/script.lua", "web-scenario:/path/to/script.lua"},
		{"_examples/github-status.lua", "web-scenario:_examples/github-status.lua"},
		{"./_examples/github-status.lua", "web-scenario:./_examples/github-status.lua"},
	}

	if runtime.GOOS == "windows" {
		tests = append(tests, Test{".\\_examples\\github-status.lua", "web-scenario:./_examples/github-status.lua"})
	}

	for _, tt := range tests {
		if u, err := ParseTargetURL(tt.input); err != nil {
			t.Errorf("%s: %s", tt.input, err)
		} else if u.String() != tt.want {
			t.Errorf("%s: expected %q but got %q", tt.input, tt.want, u.String())
		}
	}
}
