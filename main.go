package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/macrat/ayd/lib-ayd"
	"github.com/yuin/gopher-lua"
)

func asArray(t *lua.LTable) ([]interface{}, bool) {
	isArray := true
	values := make(map[int]lua.LValue)
	t.ForEach(func(k, v lua.LValue) {
		if n, ok := k.(lua.LNumber); ok {
			if math.Mod(float64(n), 1) != 0 {
				isArray = false
			} else {
				values[int(n)] = v
			}
		} else {
			isArray = false
		}
	})
	if !isArray {
		return nil, false
	}
	result := make([]interface{}, len(values))
	for i := 1; i <= len(values); i++ {
		v, ok := values[i]
		if !ok {
			return nil, false
		}
		result[i-1] = UnpackLValue(v)
	}
	return result, true
}

func UnpackLValue(v lua.LValue) interface{} {
	switch x := v.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return x == lua.LTrue
	case lua.LNumber:
		return float64(x)
	case lua.LString:
		return string(x)
	case *lua.LTable:
		if array, ok := asArray(x); ok {
			return array
		}

		values := make(map[string]interface{})
		x.ForEach(func(k, v lua.LValue) {
			values[k.String()] = UnpackLValue(v)
		})
		return values
	default:
		return x.String()
	}
}

func StartBrowser(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := chromedp.NewContext(ctx,
		chromedp.WithLogf(func(s string, args ...interface{}) {
			fmt.Printf("log"+s+"\n", args)
		}),
		//chromedp.WithDebugf(func(s string, args ...interface{}) {
		//	fmt.Printf("debug" + s + "\n", args)
		//}),
		chromedp.WithErrorf(func(s string, args ...interface{}) {
			fmt.Printf("error"+s+"\n", args)
		}),
	)
	chromedp.Run(ctx)
	return ctx, cancel
}

func NewContext() (context.Context, context.CancelFunc) {
	ctx, cancel1 := context.WithTimeout(context.Background(), time.Hour)
	//ctx, cancel2 := chromedp.NewExecAllocator(ctx, chromedp.NoFirstRun, chromedp.NoDefaultBrowserCheck)
	ctx, cancel3 := StartBrowser(ctx)

	return ctx, func() {
		cancel3()
		//cancel2()
		cancel1()
	}
}

func NewLuaState(ctx context.Context) *lua.LState {
	L := lua.NewState()

	RegisterElementType(ctx, L)
	RegisterTabType(ctx, L)
	RegisterTime(L)

	return L
}

func main() {
	args, err := ayd.ParseProbePluginArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, "$ ayd-web-script-probe TARGERT_URL")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger := ayd.NewLogger(args.TargetURL)

	ctx, cancel := NewContext()
	defer cancel()

	L := NewLuaState(ctx)
	defer L.Close()

	err = L.DoFile(args.TargetURL.Opaque)
	if err != nil {
		logger.Failure(err.Error(), nil)
		return
	}

	logger.Healthy("", nil)
}
