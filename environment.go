package main

import (
	"context"

	"github.com/yuin/gopher-lua"
)

type Environment struct {
	L       *lua.LState
	tabs    []*Tab
	logger  *Logger
	storage *Storage

	EnableRecording bool
}

func NewEnvironment(ctx context.Context, logger *Logger, s *Storage) *Environment {
	L := lua.NewState()

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
	return nil
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
