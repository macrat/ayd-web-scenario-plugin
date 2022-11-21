package webscenario

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/macrat/ayd-web-scenario/internal/lua"
)

type Storage struct {
	sync.Mutex

	Dir       string
	artifacts []string
	guids     map[string]string
	autoid    int
}

func NewStorage(baseDir string, timestamp time.Time) (*Storage, error) {
	dir := filepath.Join(baseDir, timestamp.Format("20060102T150405"))

	if !filepath.IsAbs(dir) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dir = filepath.Join(cwd, dir)
	}

	return &Storage{
		Dir:   dir,
		guids: make(map[string]string),
	}, nil
}

func (s *Storage) mkdir(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil && errors.Is(err, os.ErrExist) {
		return err
	}
	return nil
}

func (s *Storage) Open(name string) (*os.File, error) {
	p := filepath.Join(s.Dir, name)

	s.Lock()
	s.artifacts = append(s.artifacts, p)
	s.Unlock()

	if err := s.mkdir(p); err != nil {
		return nil, err
	}
	return os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
}

func (s *Storage) Remove(path string) error {
	s.Lock()
	for i, p := range s.artifacts {
		if p == path {
			s.artifacts = append(s.artifacts[:i], s.artifacts[i+1:]...)
			s.Unlock()
			return os.Remove(path)
		}
	}
	s.Unlock()
	return os.ErrNotExist
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

	if err := s.mkdir(p); err != nil {
		return err
	}
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

func (s *Storage) appendArtifact(path string) {
	for _, p := range s.artifacts {
		if p == path {
			return
		}
	}
	s.artifacts = append(s.artifacts, path)
}

func (s *Storage) CompleteDownload(guid string) string {
	s.Lock()
	defer s.Unlock()
	if name, ok := s.guids[guid]; ok {
		p := filepath.Join(s.Dir, name)
		s.appendArtifact(p)
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

func (s *Storage) Register(env *Environment, L *lua.State) {
	L.CreateTable(0, 2)

	L.SetFuncs(-1, map[string]lua.GFunction{
		"open": func(L *lua.State) int {
			env.Yield()

			path := filepath.Join(s.Dir, L.CheckString(1))
			mode := L.ToString(2)
			if mode == "" {
				mode = "r"
			}
			L.SetTop(0)

			if !strings.HasPrefix(mode, "r") {
				if err := s.mkdir(path); err != nil {
					L.Error(1, err)
				}
				s.Lock()
				s.appendArtifact(path)
				s.Unlock()
			}

			L.GetGlobal("io")
			L.GetField(-1, "open")
			L.Remove(1)

			L.PushString(path)
			L.PushString(mode)
			if err := L.Call(2, 2); err != nil {
				L.Error(1, err)
			}
			return 2
		},
		"remove": func(L *lua.State) int {
			env.Yield()

			path := filepath.Join(s.Dir, L.CheckString(1))
			if err := s.Remove(path); err != nil {
				L.Error(1, err)
			}

			return 0
		},
	})

	L.CreateTable(0, 1)
	L.SetFunction(-1, "__index", func(L *lua.State) int {
		env.Yield()

		switch L.CheckString(2) {
		case "path":
			L.PushString(s.Dir)
			return 1
		case "list":
			ls := s.Artifacts()
			L.CreateTable(len(ls), 0)
			for i, x := range ls {
				p, err := filepath.Rel(s.Dir, x)
				if err == nil {
					L.PushString(p)
				} else {
					L.PushString(x)
				}
				L.SetI(-2, i+1)
			}
			return 1
		}
		return 0
	})
	L.SetMetatable(-2)

	L.SetGlobal("artifact")
}
