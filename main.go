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

func StartBrowser(ctx context.Context, debuglog *ayd.Logger) (context.Context, context.CancelFunc) {
	var opts []chromedp.ContextOption
	if debuglog != nil {
		opts = append(
			opts,
			chromedp.WithLogf(func(s string, args ...any) {
				debuglog.Healthy(fmt.Sprintf(s, args), map[string]any{
					"level": "log",
				})
			}),
			chromedp.WithDebugf(func(s string, args ...any) {
				debuglog.Healthy(fmt.Sprintf(s, args), map[string]any{
					"level": "debug",
				})
			}),
			chromedp.WithErrorf(func(s string, args ...any) {
				debuglog.Failure(fmt.Sprintf(s, args), map[string]any{
					"level": "error",
				})
			}),
		)
	}
	ctx, cancel := chromedp.NewContext(ctx, opts...)
	chromedp.Run(ctx)
	return ctx, cancel
}

func NewContext(debuglog *ayd.Logger) (context.Context, context.CancelFunc) {
	ctx, cancel1 := context.WithTimeout(context.Background(), time.Hour)
	ctx, cancel2 := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)
	ctx, cancel3 := StartBrowser(ctx, debuglog)

	return ctx, func() {
		cancel3()
		cancel2()
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
	timestamp := time.Now()

	logger := &Logger{Debug: debug, Status: ayd.StatusHealthy}

	baseDir := os.Getenv("WEBSCENARIO_ARTIFACT_DIR")
	storage, err := NewStorage(baseDir, target.Opaque, timestamp)
	if err != nil {
		return ayd.Record{
			Time:    timestamp,
			Status:  ayd.StatusFailure,
			Message: err.Error(),
		}
	}

	var browserlog *ayd.Logger
	if debug {
		f, err := storage.Open("browser.log")
		if err != nil {
			return ayd.Record{
				Time:    timestamp,
				Status:  ayd.StatusFailure,
				Message: err.Error(),
			}
		}
		defer f.Close()
		l := ayd.NewLoggerWithWriter(f, target)
		browserlog = &l
	}
	ctx, cancel := NewContext(browserlog)
	defer cancel()

	L := NewLuaState(ctx, logger, storage)
	defer L.Close()

	stime := time.Now()
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
	r.Time = timestamp
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
