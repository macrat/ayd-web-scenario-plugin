package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_downloadFile(t *testing.T) {
	t.Setenv("TZ", "UTC")

	server := StartTestServer()
	defer server.Close()

	ctx, cancel := NewContext()
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	timestamp := time.Now()
	tmpdir := t.TempDir()

	s, err := NewStorage(tmpdir, "dummy-file.lua", timestamp)
	if err != nil {
		t.Fatalf("failed to prepare storage: %s", err)
	}

	logger := &Logger{Debug: true}
	L := NewLuaState(ctx, logger, s)
	RegisterTestUtil(L, server)

	if err := L.DoString(`tab.new(TEST.url("/download"))("a"):click()`); err != nil {
		t.Fatalf(err.Error())
	}

	for len(s.Artifacts()) == 0 {
		time.Sleep(5 * time.Millisecond)
		select {
		case <-ctx.Done():
			break
		default:
		}
	}

	b, err := os.ReadFile(filepath.Join(tmpdir, "dummy-file", timestamp.Format("20060102T150405"), "data.txt"))
	if err != nil {
		t.Log(filepath.Glob(filepath.Join(tmpdir, "dummy-file", timestamp.Format("20060102T150405"), "*")))
		t.Fatalf("failed to read downloaded file: %s", err)
	}
	if string(b) != "this is a data" {
		t.Fatalf("unexpected data has downloaded: %q", string(b))
	}
}
