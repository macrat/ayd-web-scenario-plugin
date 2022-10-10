package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/macrat/ayd/lib-ayd"
	"github.com/yuin/gopher-lua"
)

func StartBrowser(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := chromedp.NewContext(ctx,
		chromedp.WithLogf(func(s string, args ...any) {
			fmt.Printf("log"+s+"\n", args)
		}),
		//chromedp.WithDebugf(func(s string, args ...any) {
		//	fmt.Printf("debug" + s + "\n", args)
		//}),
		chromedp.WithErrorf(func(s string, args ...any) {
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

func NewLuaState(ctx context.Context, logger *Logger, s *Storage) *lua.LState {
	L := lua.NewState()

	RegisterLogger(L, logger)
	RegisterElementsArrayType(ctx, L)
	RegisterElementType(ctx, L)
	RegisterTabType(ctx, L, s)
	RegisterTime(L)

	return L
}

func RunWebScenario(target *ayd.URL, debug bool) ayd.Record {
	logger := &Logger{Debug: debug, Status: ayd.StatusHealthy}

	ctx, cancel := NewContext()
	defer cancel()

	baseDir := os.Getenv("WEBSCENARIO_ARTIFACT_DIR")

	stime := time.Now()

	storage, err := NewStorage(baseDir, target.Opaque, stime)
	if err != nil {
		return ayd.Record{
			Status:  ayd.StatusFailure,
			Message: err.Error(),
		}
	}
	L := NewLuaState(ctx, logger, storage)
	defer L.Close()

	err = L.DoFile(target.Opaque)
	latency := time.Since(stime)

	if err != nil {
		var apierr *lua.ApiError
		if errors.As(err, &apierr) {
			logger.SetExtra("error", apierr.Object.String())
			logger.SetExtra("trace", apierr.StackTrace)
		} else {
			logger.SetExtra("error", err.Error())
		}
		logger.Status = ayd.StatusFailure
	}

	if xs := storage.Artifacts(); len(xs) > 0 {
		logger.SetExtra("artifacts", xs)
	}

	r := logger.AsRecord()
	r.Time = stime
	r.Latency = latency
	return r
}

func main() {
	args, err := ayd.ParseProbePluginArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, "$ ayd-web-script-probe TARGERT_URL")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	r := RunWebScenario(args.TargetURL, true)
	ayd.NewLogger(args.TargetURL).Print(r)
}
