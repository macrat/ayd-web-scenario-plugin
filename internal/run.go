package webscenario

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/macrat/ayd/lib-ayd"
	"github.com/yuin/gopher-lua"
)

func NewContext(timeout time.Duration, debuglog *ayd.Logger) (context.Context, context.CancelFunc) {
	ctx, stopTimeout := context.WithTimeout(context.Background(), timeout)
	ctx, stopNotify := signal.NotifyContext(ctx, os.Interrupt)
	ctx, stopAllocator := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)

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
	ctx, stopBrowser := chromedp.NewContext(ctx, opts...)

	return ctx, func() {
		stopBrowser()
		stopAllocator()
		stopNotify()
		stopTimeout()
	}
}

func Run(target *ayd.URL, timeout time.Duration, debug bool, enableRecording bool) ayd.Record {
	timestamp := time.Now()

	logger := &Logger{Status: ayd.StatusHealthy}
	if debug {
		logger.DebugOut = os.Stderr
	}

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

	ctx, cancel := NewContext(timeout, browserlog)
	defer cancel()

	env := NewEnvironment(ctx, logger, storage)
	env.EnableRecording = enableRecording

	stime := time.Now()
	err = env.DoFile(target.Opaque)
	latency := time.Since(stime)

	env.Close()

	if err != nil {
		logger.Status = ayd.StatusFailure

		var apierr *lua.ApiError
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.SetExtra("error", "timeout")
			logger.Status = ayd.StatusAborted
		} else if errors.Is(ctx.Err(), context.Canceled) {
			logger.SetExtra("error", "interrupted")
			logger.Status = ayd.StatusAborted
		} else if errors.As(err, &apierr) {
			logger.SetExtra("error", apierr.Object.String())
			logger.SetExtra("trace", apierr.StackTrace)
		} else {
			logger.SetExtra("error", err.Error())
		}
	}

	if xs := storage.Artifacts(); len(xs) > 0 {
		logger.SetExtra("artifacts", xs)
	}

	r := logger.AsRecord()
	r.Time = timestamp
	r.Latency = latency
	return r
}
