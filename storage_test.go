package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNewStorage(t *testing.T) {
	t.Setenv("TZ", "UTC")

	timestamp := time.Date(2022, 1, 2, 15, 4, 5, 0, time.UTC)
	tmpdir := t.TempDir()
	script := filepath.Join(tmpdir, "file.name.lua")

	if s, err := NewStorage("", script, timestamp); err != nil {
		t.Errorf("failed to create storage: %s", err)
	} else if s.Dir != filepath.Join(tmpdir, "file.name", "20220102T150405") {
		t.Errorf("unexpected storage directory: %s", s.Dir)
	}

	if s, err := NewStorage(tmpdir, script, timestamp); err != nil {
		t.Errorf("failed to create storage: %s", err)
	} else if s.Dir != filepath.Join(tmpdir, "file.name", "20220102T150405") {
		t.Errorf("unexpected storage directory: %s", s.Dir)
	}

	if s, err := NewStorage(filepath.Join(tmpdir, "output"), script, timestamp); err != nil {
		t.Errorf("failed to create storage: %s", err)
	} else if s.Dir != filepath.Join(tmpdir, "output", "file.name", "20220102T150405") {
		t.Errorf("unexpected storage directory: %s", s.Dir)
	}
}
