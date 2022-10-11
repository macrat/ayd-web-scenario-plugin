package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Storage struct {
	sync.Mutex

	Dir       string
	artifacts []string
	guids     map[string]string
	autoid    int
}

func NewStorage(basedir, scriptPath string, timestamp time.Time) (*Storage, error) {
	if basedir == "" {
		basedir = filepath.Dir(scriptPath)
	}

	name := filepath.Base(scriptPath)
	name = name[:len(name)-len(filepath.Ext(name))]
	dir := filepath.Join(basedir, name, timestamp.Format("20060102T150405"))

	if err := os.MkdirAll(dir, 0750); err != nil && errors.Is(err, os.ErrExist) {
		return nil, err
	}

	return &Storage{
		Dir:   dir,
		guids: make(map[string]string),
	}, nil
}

func (s *Storage) Save(name, ext string, data []byte) error {
	s.Lock()

	if name == "" {
		s.autoid += 1
		name = fmt.Sprintf("%06d", s.autoid)
	}
	if !strings.HasSuffix(name, ext) {
		name += ext
	}

	p := filepath.Join(s.Dir, name)
	s.artifacts = append(s.artifacts, p)

	s.Unlock()

	return os.WriteFile(p, data, 0644)
}

func (s *Storage) StartDownload(guid, name string) {
	s.Lock()
	defer s.Unlock()
	s.guids[guid] = name
}

func (s *Storage) CancelDownload(guid string) {
	s.Lock()
	defer s.Unlock()
	delete(s.guids, guid)
}

func (s *Storage) CompleteDownload(guid string) string {
	s.Lock()
	defer s.Unlock()
	if name, ok := s.guids[guid]; ok {
		p := filepath.Join(s.Dir, name)
		s.artifacts = append(s.artifacts, p)
		delete(s.guids, guid)
		return p
	}
	return ""
}

func (s *Storage) Artifacts() []string {
	s.Lock()
	defer s.Unlock()

	return append(make([]string, 0, len(s.artifacts)), s.artifacts...)
}
