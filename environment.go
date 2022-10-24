package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/yuin/gopher-lua"
)

type Environment struct {
	sync.Mutex // this mutex works like the GIL in Python.

	lua      *lua.LState
	tabs     []*Tab
	logger   *Logger
	storage  *Storage
	saveWG   sync.WaitGroup
	recordID atomic.Int64

	EnableRecording bool
}

func NewEnvironment(ctx context.Context, logger *Logger, s *Storage) *Environment {
	L := lua.NewState()
	L.SetContext(ctx)

	env := &Environment{
		lua:     L,
		logger:  logger,
		storage: s,
	}
	env.Lock()

	RegisterLogger(L, logger)
	RegisterElementType(ctx, L)
	RegisterTabType(ctx, env)
	RegisterTime(ctx, env)

	return env
}

func (env *Environment) Close() error {
	defer env.Unlock()
	for _, t := range env.tabs {
		t.Close()
	}
	env.lua.Close()
	env.saveWG.Wait()
	return nil
}

func (env *Environment) HandleError(err error) {
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			env.lua.RaiseError("timeout")
		} else if errors.Is(err, context.Canceled) {
			env.lua.RaiseError("interrupted")
		} else {
			env.lua.RaiseError("%s", err)
		}
	}
}

func (env *Environment) DoFile(path string) error {
	return env.lua.DoFile(path)
}

func (env *Environment) NewFunction(f lua.LGFunction) *lua.LFunction {
	return env.lua.NewFunction(f)
}

// Yield makes a chance to execute callback function.
func (env *Environment) Yield() {
	env.Unlock()
	env.Lock()
}

// AsyncRun makes a chance to execute callback function while executing a heavy function.
func AsyncRun[T any](env *Environment, f func() T) T {
	env.Unlock()
	defer env.Lock()
	return f()
}

// CallEventHandler calls an event callback function with GIL.
func (env *Environment) CallEventHandler(f *lua.LFunction, args map[string]lua.LValue, nret int) []lua.LValue {
	env.Lock()
	defer env.Unlock()

	L, cancel := env.lua.NewThread()
	defer cancel()

	arg := L.NewTable()
	for k, v := range args {
		L.SetField(arg, k, v)
	}

	L.Push(f)
	L.Push(arg)
	L.Call(1, nret)

	var result []lua.LValue
	for i := 1; i <= nret; i++ {
		result = append(result, L.Get(i))
	}
	return result
}

func (env *Environment) StartTask(where, taskName string) {
	env.logger.StartTask(where, taskName)
}

func (env *Environment) RegisterNewType(name string, methods map[string]lua.LGFunction, fields map[string]lua.LValue) {
	tbl := env.lua.SetFuncs(env.lua.NewTypeMetatable(name), methods)
	for k, v := range fields {
		env.lua.SetField(tbl, k, v)
	}
	env.lua.SetGlobal(name, tbl)
}

func (env *Environment) saveRecord(recorder *Recorder) {
	env.saveWG.Add(1)
	id := env.recordID.Add(1)
	go func(id int64) {
		<-recorder.Done
		if f, err := env.storage.Open(fmt.Sprintf("record-%04d.gif", id)); err == nil {
			defer f.Close()
			recorder.SaveTo(f)
		}
		env.saveWG.Done()
	}(id)
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
