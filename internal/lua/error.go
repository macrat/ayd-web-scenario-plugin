package lua

import (
	"fmt"
)

type Error struct {
	Err         error
	ChunkName   string
	CurrentLine int
	Traceback   string
}

func (err Error) OneLine() string {
	return fmt.Sprintf("%s:%d: %s", err.ChunkName, err.CurrentLine, err.Err)
}

func (err Error) Error() string {
	return err.OneLine() + "\n" + err.Traceback
}

func (err Error) Unwrap() error {
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
		if e, ok := err.(Error); ok {
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
