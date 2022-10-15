package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/macrat/ayd/lib-ayd"
)

func TestRunWebScenario(t *testing.T) {
	server := StartTestServer()
	defer server.Close()

	t.Setenv("TZ", "UTC")
	t.Setenv("TEST_URL", server.URL)
	t.Setenv("WEBSCENARIO_ARTIFACT_DIR", t.TempDir())

	tests := []struct {
		Status string
		Text   string
		Extra  string
		Error  string
		Record ayd.Record
	}{
		{
			Text: "world",
			Record: ayd.Record{
				Status:  ayd.StatusHealthy,
				Message: "It's working!",
				Extra:   nil,
			},
		},
		{
			Status: "degrade",
			Text:   "world",
			Record: ayd.Record{
				Status:  ayd.StatusDegrade,
				Message: "It's working!",
				Extra:   nil,
			},
		},
		{
			Status: "UNKNOWN",
			Text:   "world",
			Record: ayd.Record{
				Status:  ayd.StatusUnknown,
				Message: "It's working!",
				Extra:   nil,
			},
		},
		{
			Text: "incorrect",
			Record: ayd.Record{
				Status:  ayd.StatusFailure,
				Message: "It's working!",
				Extra: map[string]any{
					"error": "./testdata/_main-test.lua:13: assertion failed!",
					"trace": "stack traceback:\n\t[G]: in function 'assert'\n\t./testdata/_main-test.lua:13: in main chunk\n\t[G]: ?",
				},
			},
		},
		{
			Status: "degrade",
			Text:   "incorrect",
			Extra:  "hello",
			Record: ayd.Record{
				Status:  ayd.StatusFailure,
				Message: "It's working!",
				Extra: map[string]any{
					"msg":   "hello",
					"error": "./testdata/_main-test.lua:13: assertion failed!",
					"trace": "stack traceback:\n\t[G]: in function 'assert'\n\t./testdata/_main-test.lua:13: in main chunk\n\t[G]: ?",
				},
			},
		},
		{
			Text:  "world",
			Extra: "hello",
			Record: ayd.Record{
				Status:  ayd.StatusHealthy,
				Message: "It's working!",
				Extra: map[string]any{
					"msg": "hello",
				},
			},
		},
		{
			Text:  "world",
			Error: "something",
			Record: ayd.Record{
				Status:  ayd.StatusFailure,
				Message: "It's working!",
				Extra: map[string]any{
					"error": "./testdata/_main-test.lua:16: something",
					"trace": "stack traceback:\n\t[G]: in function 'error'\n\t./testdata/_main-test.lua:16: in main chunk\n\t[G]: ?",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status=%s/text=%s/extra=%s/error=%s", tt.Status, tt.Text, tt.Extra, tt.Error), func(t *testing.T) {
			t.Setenv("TEST_STATUS", tt.Status)
			t.Setenv("TEST_TEXT", tt.Text)
			t.Setenv("TEST_EXTRA", tt.Extra)
			t.Setenv("TEST_ERROR", tt.Error)

			var r ayd.Record

			RunWebScenario(&ayd.URL{Scheme: "web-scenario", Opaque: "./testdata/_main-test.lua"}, false, false, func(rec ayd.Record) {
				r = rec
			})

			if r.Latency == 0 {
				t.Errorf("latency should not be 0")
			}

			if r.Status != tt.Record.Status {
				t.Errorf("expected status is %s but got %s", tt.Record.Status, r.Status)
			}

			if r.Message != tt.Record.Message {
				t.Errorf("expected message is %q but got %q", tt.Record.Message, r.Message)
			}

			if diff := cmp.Diff(r.Extra, tt.Record.Extra); diff != "" {
				t.Errorf("unexpected extra:\n%s", diff)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"web-scenario:/path/to/script.lua", "web-scenario:/path/to/script.lua"},
		{"web-scenario:./path/to/script.lua", "web-scenario:./path/to/script.lua"},
		{"web-scenario:path/to/script.lua", "web-scenario:path/to/script.lua"},
		{"web-scenario:///path/to/script.lua", "web-scenario:/path/to/script.lua"},
		{"web-scenario://localhost/path/to/script.lua", "web-scenario:/path/to/script.lua"},
		{"_examples/github-status.lua", "web-scenario:_examples/github-status.lua"},
		{"./_examples/github-status.lua", "web-scenario:./_examples/github-status.lua"},
	}

	for _, tt := range tests {
		if u, err := ParseTargetURL(tt.input); err != nil {
			t.Errorf("%s: %s", tt.input, err)
		} else if u.String() != tt.want {
			t.Errorf("%s: expected %q but got %q", tt.input, tt.want, u.String())
		}
	}
}
