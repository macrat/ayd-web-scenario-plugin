package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/yuin/gopher-lua"
)

type Environment struct {
	L        *lua.LState
	tabs     []*Tab
	logger   *Logger
	storage  *Storage
	saveWG   sync.WaitGroup
	recordID int

	EnableRecording bool
}

func NewEnvironment(ctx context.Context, logger *Logger, s *Storage) *Environment {
	L := lua.NewState()
	L.SetContext(ctx)

	env := &Environment{
		L:       L,
		logger:  logger,
		storage: s,
	}

	RegisterLogger(L, logger)
	RegisterElementsArrayType(ctx, L)
	RegisterElementType(ctx, L)
	RegisterTabType(ctx, env)
	RegisterTime(L)

	return env
}

func (env *Environment) Close() error {
	for _, t := range env.tabs {
		t.Close(env.L)
	}
	env.L.Close()
	env.saveWG.Wait()
	return nil
}

func (env *Environment) saveRecord(recorder *Recorder) {
	env.saveWG.Add(1)
	env.recordID += 1 // TODO: make it thread safe?
	go func(id int) {
		<-recorder.Done
		if f, err := env.storage.Open(fmt.Sprintf("record-%04d.gif", id)); err == nil {
			defer f.Close()
			recorder.SaveTo(f)
		}
		env.saveWG.Done()
	}(env.recordID)
}

func (env *Environment) registerTab(t *Tab) {
	env.tabs = append(env.tabs, t)
}

func (env *Environment) unregisterTab(t *Tab) {
	tabs := make([]*Tab, 0, len(env.tabs))
	for _, x := range env.tabs {
		if x != t {
			tabs = append(tabs, x)
		}
	}
	env.tabs = tabs
}
