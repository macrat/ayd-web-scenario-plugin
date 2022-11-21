package lua

import (
	"errors"
	"fmt"
)

var (
	ErrRuntime = errors.New("runtime error")
	ErrMemory  = errors.New("memory allocation error")
	ErrError   = errors.New("error")
	ErrSyntax  = errors.New("syntax error")
	ErrYield   = errors.New("yield")
	ErrFile    = errors.New("file error")
	ErrUnknown = errors.New("unknown error")
)

// LuaError is raw error from Lua.
type LuaError struct {
	Kind    error
	Message string
}

func (e LuaError) Unwrap() error {
	return e.Kind
}

func (e LuaError) Error() string {
	return e.Message
}

// ErrorWithTrace is an error with traceback.
type ErrorWithTrace struct {
	Err         error
	ChunkName   string
	CurrentLine int
	Traceback   string
}

func (err ErrorWithTrace) OneLine() string {
	return fmt.Sprintf("%s:%d: %s", err.ChunkName, err.CurrentLine, err.Err)
}

func (err ErrorWithTrace) Error() string {
	return err.OneLine() + "\n" + err.Traceback
}

func (err ErrorWithTrace) Unwrap() error {
	return err.Err
}

func pcall(L *State) int {
	err := L.Call(L.GetTop()-1, MultRet)
	if err == nil {
		L.PushBoolean(true)
		L.Rotate(1, 1)
		return L.GetTop()
	} else {
		L.PushBoolean(false)
		if e, ok := err.(ErrorWithTrace); ok {
			L.PushString(e.OneLine())
		} else {
			L.PushString(err.Error())
		}
		return 2
	}
}

func xpcall(L *State) int {
	L.Swap(1, 2)

	err := L.Call(L.GetTop()-2, MultRet)

	if err == nil {
		L.PushBoolean(true)
		L.Replace(1)
		return L.GetTop()
	} else {
		if err := L.Call(1, MultRet); err != nil {
			L.Error(1, err)
		}
		L.PushBoolean(false)
		L.Rotate(1, 1)
		return L.GetTop()
	}
}
