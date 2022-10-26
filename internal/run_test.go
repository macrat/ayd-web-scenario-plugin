package webscenario

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/macrat/ayd/lib-ayd"
)

func TestRun(t *testing.T) {
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

			r := Run(&ayd.URL{Scheme: "web-scenario", Opaque: "./testdata/_main-test.lua"}, 5*time.Minute, false, false)

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

	t.Run("timeout", func(t *testing.T) {
		r := Run(&ayd.URL{Scheme: "web-scenario", Opaque: "./testdata/_timeout-test.lua"}, 100*time.Millisecond, false, false)

		if r.Status != ayd.StatusAborted {
			t.Errorf("expected ABORTED status but got %s", r.Status)
		}

		if r.Latency > 500*time.Millisecond {
			t.Errorf("unexpected latency: %s", r.Latency)
		}

		if r.Message != "I'm gonna be timeout" {
			t.Errorf("unexpected message: %q", r.Message)
		}
	})
}
