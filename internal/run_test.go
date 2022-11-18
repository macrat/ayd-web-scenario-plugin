package webscenario

import (
	"fmt"
	"testing"
	"time"
	"strings"

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
		Status  string
		Latency string
		Text    string
		Extra   string
		Error   string
		Record  ayd.Record
	}{
		{
			Text:    "world",
			Latency: "0",
			Record: ayd.Record{
				Status:  ayd.StatusHealthy,
				Message: "It's working!",
			},
		},
		{
			Status:  "degrade",
			Latency: "123",
			Text:    "world",
			Record: ayd.Record{
				Status:  ayd.StatusDegrade,
				Message: "It's working!",
				Latency: 123 * time.Millisecond,
			},
		},
		{
			Status:  "UNKNOWN",
			Latency: "1",
			Text:    "world",
			Record: ayd.Record{
				Status:  ayd.StatusUnknown,
				Latency: 1 * time.Millisecond,
				Message: "It's working!",
			},
		},
		{
			Text:    "incorrect",
			Latency: "0",
			Record: ayd.Record{
				Status:  ayd.StatusFailure,
				Message: strings.Join([]string{
					`It's working!`,
					`./testdata/run-test.lua:17: assertion failed: "world" == "incorrect"`,
					`stack traceback:`,
					`	[G]: in function 'eq'`,
					`	./testdata/run-test.lua:17: in main chunk`,
					`	[G]: ?`,
				}, "\n"),
			},
		},
		{
			Status:  "degrade",
			Latency: "0",
			Text:    "incorrect",
			Extra:   "hello",
			Record: ayd.Record{
				Status:  ayd.StatusFailure,
				Message: strings.Join([]string{
					`It's working!`,
					`./testdata/run-test.lua:17: assertion failed: "world" == "incorrect"`,
					`stack traceback:`,
					`	[G]: in function 'eq'`,
					`	./testdata/run-test.lua:17: in main chunk`,
					`	[G]: ?`,
				}, "\n"),
				Extra: map[string]any{
					"msg": "hello",
				},
			},
		},
		{
			Latency: "0",
			Text:    "world",
			Extra:   "hello",
			Record: ayd.Record{
				Status:  ayd.StatusHealthy,
				Message: "It's working!",
				Extra: map[string]any{
					"msg": "hello",
				},
			},
		},
		{
			Latency: "0",
			Text:    "world",
			Error:   "something",
			Record: ayd.Record{
				Status:  ayd.StatusFailure,
				Message: strings.Join([]string{
					`It's working!`,
					`./testdata/run-test.lua:20: something`,
					`stack traceback:`,
					`	[G]: in function 'error'`,
					`	./testdata/run-test.lua:20: in main chunk`,
					`	[G]: ?`,
				}, "\n"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status=%s/text=%s/extra=%s/error=%s", tt.Status, tt.Text, tt.Extra, tt.Error), func(t *testing.T) {
			t.Setenv("TEST_STATUS", tt.Status)
			t.Setenv("TEST_LATENCY", tt.Latency)
			t.Setenv("TEST_TEXT", tt.Text)
			t.Setenv("TEST_EXTRA", tt.Extra)
			t.Setenv("TEST_ERROR", tt.Error)

			r := Run(Arg{
				Mode:    "ayd",
				Target:  &ayd.URL{Scheme: "web-scenario", Opaque: "./testdata/run-test.lua"},
				Timeout: 5 * time.Minute,
			})

			if (r.Latency - tt.Record.Latency).Truncate(time.Millisecond) != 0 {
				t.Errorf("expected latency is %s but got %s", tt.Record.Latency, r.Latency)
			}

			if r.Status != tt.Record.Status {
				t.Errorf("expected status is %s but got %s", tt.Record.Status, r.Status)
			}

			if r.Message != tt.Record.Message {
				t.Errorf("expected message is %q but got %q", tt.Record.Message, r.Message)
			}

			if diff := cmp.Diff(tt.Record.Extra, r.Extra); diff != "" {
				t.Errorf("unexpected extra:\n%s", diff)
			}
		})
	}

	t.Run("timeout", func(t *testing.T) {
		r := Run(Arg{
			Mode:    "ayd",
			Target:  &ayd.URL{Scheme: "web-scenario", Opaque: "./testdata/timeout.lua"},
			Timeout: 500 * time.Millisecond,
		})

		if r.Status != ayd.StatusUnknown {
			t.Errorf("expected FAILURE status but got %s", r.Status)
		}

		if r.Latency > 1*time.Second {
			t.Errorf("unexpected latency: %s", r.Latency)
		}

		expect := strings.Join([]string{
			`I'm gonna be timeout`,
			`./testdata/timeout.lua:2: timeout`,
			`stack traceback:`,
			`	[G]: in function 'sleep'`,
			`	./testdata/timeout.lua:2: in main chunk`,
			`	[G]: ?`,
		}, "\n")
		if r.Message != expect {
			t.Errorf("unexpected message: %q", r.Message)
		}
	})
}
