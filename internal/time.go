package webscenario

import (
	"context"
	"fmt"
	"time"

	"github.com/yuin/gopher-lua"
)

func RegisterTime(ctx context.Context, env *Environment) {
	env.RegisterNewType("time", map[string]lua.LGFunction{
		"now": func(L *lua.LState) int {
			env.Yield()
			L.Push(lua.LNumber(time.Now().UnixMilli()))
			return 1
		},
		"sleep": func(L *lua.LState) int {
			n := float64(L.CheckNumber(1))
			env.RecordOnAllTabs(L, fmt.Sprintf("time.sleep(%f)", n))

			dur := time.Duration(n * float64(time.Millisecond))
			err := AsyncRun(env, func() error {
				var err error
				timer := time.NewTimer(dur)
				select {
				case <-timer.C:
					err = nil
				case <-ctx.Done():
					err = ctx.Err()
				}
				timer.Stop()
				return err
			})
			env.HandleError(err)

			env.RecordOnAllTabs(L, fmt.Sprintf("time.sleep(%f)", n))
			return 0
		},
		"format": func(L *lua.LState) int {
			env.Yield()
			L.Push(lua.LString(time.UnixMilli(int64(L.CheckNumber(1))).Format(time.RFC3339)))
			return 1
		},
	}, map[string]lua.LValue{
		"millisecond": lua.LNumber(1),
		"second":      lua.LNumber(1000),
		"minute":      lua.LNumber(60 * 1000),
		"hour":        lua.LNumber(60 * 60 * 1000),
	})
}
