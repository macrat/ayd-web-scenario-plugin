package webscenario

import (
	"context"
	"fmt"
	"time"

	"github.com/macrat/ayd-web-scenario/internal/lua"
)

func RegisterTime(ctx context.Context, env *Environment, L *lua.State) {
	L.CreateTable(0, 0)
	L.SetFuncs(-1, map[string]lua.GFunction{
		"now": func(L *lua.State) int {
			env.Yield()
			L.PushInteger(time.Now().UnixMilli())
			return 1
		},
		"sleep": func(L *lua.State) int {
			n := float64(L.CheckNumber(1))
			env.RecordOnAllTabs(L, fmt.Sprintf("time.sleep(%f)", n))

			dur := time.Duration(n * float64(time.Millisecond))
			AsyncRun(env, L, func() (struct{}, error) {
				var err error
				timer := time.NewTimer(dur)
				select {
				case <-timer.C:
					err = nil
				case <-ctx.Done():
					err = ctx.Err()
				}
				timer.Stop()
				return struct{}{}, err
			})

			env.RecordOnAllTabs(L, fmt.Sprintf("time.sleep(%f)", n))
			return 0
		},
		"format": func(L *lua.State) int {
			env.Yield()

			n := L.CheckNumber(1)
			format := "%Y-%m-%dT%H:%M:%S%z"
			if L.Type(2) != lua.Nil {
				L.ToString(2)
			}

			L.GetGlobal("os")
			L.GetField(-1, "date")
			L.PushString(format)
			L.PushNumber(n / 1000)
			if err := L.Call(2, 1); err != nil {
				L.Error(1, err)
			}
			return 1
		},
	})
	L.SetInteger(-1, "millisecond", 1)
	L.SetInteger(-1, "second", 1000)
	L.SetInteger(-1, "minute", 1000*60)
	L.SetInteger(-1, "hour", 1000*60*60)
	L.SetInteger(-1, "day", 1000*60*60*24)
	L.SetInteger(-1, "week", 1000*60*60*24*7)
	L.SetInteger(-1, "year", 1000*60*60*24*7*365)
	L.SetGlobal("time")
}
