package main

import (
	"runtime"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	type Test struct {
		input string
		mode  string
		want  string
	}

	tests := []Test{
		{"web-scenario:/path/to/script.lua", "ayd", "web-scenario:/path/to/script.lua"},
		{"web-scenario:./path/to/script.lua", "ayd", "web-scenario:./path/to/script.lua"},
		{"web-scenario:path/to/script.lua", "ayd", "web-scenario:path/to/script.lua"},
		{"web-scenario:///path/to/script.lua", "ayd", "web-scenario:/path/to/script.lua"},
		{"web-scenario://localhost/path/to/script.lua", "ayd", "web-scenario:/path/to/script.lua"},
		{"web-scenario://foo:bar@localhost/path/to/script.lua", "ayd", "web-scenario://foo:xxxxx@/path/to/script.lua"},
		{"web-scenario://foo:bar@/path/to/script.lua", "ayd", "web-scenario://foo:xxxxx@/path/to/script.lua"},
		{"web-scenario://foo@/path/to/script.lua", "ayd", "web-scenario://foo@/path/to/script.lua"},
		{"examples/github-status.lua", "standalone", "web-scenario:examples/github-status.lua"},
		{"./examples/github-status.lua", "standalone", "web-scenario:./examples/github-status.lua"},
		{"examples/github-status.lua", "standalone", "web-scenario:examples/github-status.lua"},
	}

	if runtime.GOOS == "windows" {
		tests = append(
			tests,
			Test{`web-scenario://foo:bar@/C:\path\to\script.lua`, "ayd", "web-scenario:./examples/github-status.lua"},
			Test{`.\examples\github-status.lua`, "standalone", "web-scenario:./examples/github-status.lua"},
			Test{`examples\github-status.lua`, "standalone", "web-scenario:examples/github-status.lua"},
		)
	}

	for _, tt := range tests {
		if mode, u, err := ParseTargetURL(tt.input); err != nil {
			t.Errorf("%s: %s", tt.input, err)
		} else if u.String() != tt.want {
			t.Errorf("%s: expected %q but got %q", tt.input, tt.want, u.String())
		} else if mode != tt.mode {
			t.Errorf("%s: expected mode %s but got %s", tt.input, tt.mode, mode)
		}
	}
}
