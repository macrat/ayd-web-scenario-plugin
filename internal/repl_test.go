package webscenario

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/macrat/ayd/lib-ayd"
	"github.com/yuin/gopher-lua"
)

func Test_isIncomplete(t *testing.T) {
	tests := []struct {
		Script string
		Want   bool
	}{
		{"print('hello'", true},
		{"print'hello')", false},
	}

	L := lua.NewState()
	defer L.Close()

	for _, tt := range tests {
		_, err := L.LoadString(tt.Script)
		actual := isIncomplete(err)
		if actual != tt.Want {
			t.Errorf("%s => expected %v but got %v", tt.Script, tt.Want, actual)
		}
	}
}

func TestEnvironment_DoStream(t *testing.T) {
	server := StartTestServer()
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	storage, err := NewStorage(t.TempDir(), time.Now())
	if err != nil {
		t.Fatalf("failed to prepare storage: %s", err)
	}

	var log strings.Builder
	env := NewEnvironment(ctx, &Logger{Stream: &log}, storage, Arg{Mode: "stdin", Target: &ayd.URL{Scheme: "web-scenario", Opaque: "<stdin>"}, Timeout: 5 * time.Minute})
	defer env.Close()

	tests := []struct {
		Input  string
		Log    string
		Source []string
	}{
		{
			"print('hello')\nprint('world')\n",
			"hello\nworld\n",
			[]string{"print('hello')", "print('world')", ""},
		},
		{
			"print('hello')\nprint('world')",
			"hello\nworld\n",
			[]string{"print('hello')", "print('world')", ""},
		},
		{
			"print('yo')\n",
			"yo\n",
			[]string{"print('yo')", ""},
		},
		{
			"print('yo')",
			"yo\n",
			[]string{"print('yo')", ""},
		},
		{
			"",
			"",
			[]string{},
		},
	}

	for i, tt := range tests {
		var log strings.Builder
		env.logger.Stream = &log

		sourceImager.sources = make(map[string][]string)

		err = env.DoStream(strings.NewReader(tt.Input), "stdin")
		if err != nil {
			t.Fatalf("%d: unexpected error: %s", i, err)
		}

		if log.String() != tt.Log {
			t.Errorf("%d: unexpected stdout: %q", i, log.String())
		}

		if s, ok := sourceImager.sources["<stdin>"]; !ok {
			t.Errorf("%d: source not found in the source imager", i)
		} else if diff := cmp.Diff(tt.Source, s); diff != "" {
			t.Errorf("%d: unexpected source recorded:\n%s", i, diff)
		}
	}
}
