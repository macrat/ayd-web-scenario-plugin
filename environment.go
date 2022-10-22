package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/yuin/gopher-lua"
)

type Environment struct {
	sync.Mutex // this mutex works like the GIL in Python.

	lua      *lua.LState
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
		lua:     L,
		logger:  logger,
		storage: s,
	}

	RegisterLogger(L, logger)
	RegisterElementsArrayType(ctx, L)
	RegisterElementType(ctx, L)
	RegisterTabType(ctx, env)
	RegisterTime(env)

	return env
}

func (env *Environment) Close() error {
	for _, t := range env.tabs {
		t.Close(env.lua)
	}
	env.lua.Close()
	env.saveWG.Wait()
	return nil
}

func (env *Environment) RaiseError(fmt string, args ...interface{}) {
	env.lua.RaiseError(fmt, args...)
}

func (env *Environment) DoFile(path string) error {
	env.Lock()
	defer env.Unlock()
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

// Callback calls a callback function with GIL.
func (env *Environment) Callback(f *lua.LFunction, args []lua.LValue, nret int) []lua.LValue {
	env.Lock()
	defer env.Unlock()

	L, cancel := env.lua.NewThread()
	defer cancel()

	L.CallByParam(lua.P{
		Fn:   f,
		NRet: nret,
	}, args...)

	var result []lua.LValue
	for i := 1; i <= nret; i++ {
		result = append(result, L.Get(i))
	}
	return result
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
