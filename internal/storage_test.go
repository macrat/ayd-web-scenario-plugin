package webscenario

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNewStorage(t *testing.T) {
	t.Parallel()

	timestamp := time.Date(2022, 1, 2, 15, 4, 5, 0, time.UTC)
	tmpdir := t.TempDir()

	if s, err := NewStorage(tmpdir, timestamp); err != nil {
		t.Errorf("failed to create storage: %s", err)
	} else if s.Dir != filepath.Join(tmpdir, "20220102T150405") {
		t.Errorf("unexpected storage directory: %s", s.Dir)
	}
}
