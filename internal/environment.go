package webscenario

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/macrat/ayd-web-scenario/internal/lua"
)

var (
	Version = "HEAD"
	Commit  = "UNKNOWN"
)

type Environment struct {
	sync.Mutex // this mutex works like the GIL in Python.

	lua     *lua.State
	ctx     context.Context
	stop    context.CancelFunc
	tabs    []*Tab
	logger  *Logger
	storage *Storage
	saveWG  sync.WaitGroup
	errch   chan error

	EnableRecording bool
}

func NewEnvironment(ctx context.Context, logger *Logger, s *Storage, arg Arg) (*Environment, error) {
	L, err := lua.NewState()
	if err != nil {
		return nil, err
	}

	env := &Environment{
		lua:     L,
		ctx:     ctx,
		stop:    func() {},
		logger:  logger,
		storage: s,
		errch:   make(chan error, 1),
	}
	env.Lock()

	RegisterLogger(L, logger)
	RegisterElementType(ctx, L)
	RegisterTabType(ctx, env, L)
	RegisterTime(ctx, env, L)
	RegisterAssert(L)
	RegisterKey(L)
	RegisterFileLike(L)
	RegisterEncodings(env, L)
	RegisterFetch(ctx, env, L)
	s.Register(env, L)
	arg.Register(L)

	return env, nil
}

func (env *Environment) Close() error {
	defer env.Unlock()
	for _, t := range env.tabs {
		t.Close()
	}
	env.lua.Close()
	env.stop()
	env.saveWG.Wait()
	close(env.errch)
	return nil
}

func HandleError(L *lua.State, err error) {
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			L.Errorf(1, "timeout")
		} else if errors.Is(err, context.Canceled) {
			L.Errorf(1, "interrupted")
		} else {
			L.Error(1, err)
		}
	}
}

func (env *Environment) DoFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	ctx, stop := context.WithCancel(env.ctx)
	env.stop = stop

	done := make(chan struct{})

	go func() {
		err := env.lua.DoWithContext(ctx, f, path, true)
		if err != nil {
			env.errch <- err
		}
		close(done)
	}()

	select {
	case <-done:
	case err = <-env.errch:
		env.stop()
		<-done
	}
	env.stop = func() {}
	return err
}

// Yield makes a chance to execute callback function.
func (env *Environment) Yield() {
	env.Unlock()
	env.Lock()
}

// AsyncRun makes a chance to execute callback function while executing a heavy function.
func AsyncRun[T any](env *Environment, L *lua.State, f func() (T, error)) T {
	env.Unlock()
	defer env.Lock()
	x, err := f()
	HandleError(L, err)
	return x
}

/*
// CallEventHandler calls an event callback function with GIL.
func (env *Environment) CallEventHandler(f *lua.LFunction, arg *lua.LTable, nret int) []lua.LValue {
	env.Lock()
	defer env.Unlock()

	L, cancel := env.lua.NewThread()
	defer cancel()

	L.Push(f)
	L.Push(arg)
	env.errch <- L.PCall(1, nret, nil)

	var result []lua.LValue
	for i := 1; i <= nret; i++ {
		result = append(result, L.Get(i))
	}
	return result
}
*/

func (env *Environment) StartTask(where string, line int, taskName string) {
	env.logger.StartTask(where, line, taskName)
}

func (env *Environment) BuildTable(build func(L *lua.State)) {
	env.Lock() // XXX: why is it lock here? it's weird
	defer env.Unlock()

	env.lua.CreateTable(0, 0)
	build(env.lua)
}

func (env *Environment) saveRecord(id int, recorder *Recorder) {
	env.saveWG.Add(1)
	go func(id int) {
		<-recorder.Done
		if f, err := env.storage.Open(fmt.Sprintf("record%d.gif", id)); err == nil {
			err = recorder.SaveTo(f)
			f.Close()
			if err == NoRecord {
				env.storage.Remove(f.Name())
			}
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

func (env *Environment) RecordOnAllTabs(L *lua.State, taskName string) {
	for _, tab := range env.tabs {
		tab.RecordOnce(L, taskName)
	}
}
