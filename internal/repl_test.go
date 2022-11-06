package webscenario

import (
	"time"
	"context"
	"strings"
	"testing"

	"github.com/yuin/gopher-lua"
	"github.com/macrat/ayd/lib-ayd"
	"github.com/google/go-cmp/cmp"
)

func Test_isIncomplete(t *testing.T) {
	tests := []struct{
		Script string
		Want   bool
	} {
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

	ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Minute)
	defer cancel()

	storage, err := NewStorage(t.TempDir(), time.Now())
	if err != nil {
		t.Fatalf("failed to prepare storage: %s", err)
	}

	var log strings.Builder

	env := NewEnvironment(ctx, &Logger{Stream: &log}, storage, Arg{Mode: "stdin", Target: &ayd.URL{Scheme: "web-scenario", Opaque: "<stdin>"}, Timeout: 5 * time.Minute})
	defer env.Close()

	sourceImager.sources = make(map[string][]string)

	err = env.DoStream(strings.NewReader("print('hello')\nprint('world')"), "stdin")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if log.String() != "hello\nworld\n" {
		t.Errorf("unexpected stdout: %q", log.String())
	}

	if s, ok := sourceImager.sources["<stdin>"]; !ok {
		t.Errorf("source not found in the source imager")
	} else if diff := cmp.Diff([]string{"print('hello')", "print('world')"}, s); diff != "" {
		t.Errorf("unexpected source recorded:\n%s", diff)
	}
}
