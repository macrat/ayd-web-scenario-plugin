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

func NewExecAllocator(ctx context.Context, withHead bool) (context.Context, context.CancelFunc) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,

		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-features", "site-per-process,Translate,BlinkGenPropertyTrees"),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),
	}
	if !withHead {
		opts = append(opts, chromedp.Headless)
	}
	return chromedp.NewExecAllocator(ctx, opts...)
}

func NewContext(timeout time.Duration, withHead bool, debuglog *ayd.Logger) (context.Context, context.CancelFunc) {
	ctx, stopTimeout := context.WithTimeout(context.Background(), timeout)
	ctx, stopNotify := signal.NotifyContext(ctx, os.Interrupt)
	ctx, stopAllocator := NewExecAllocator(ctx, withHead)

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

type Options struct {
	Target    *ayd.URL
	Timeout   time.Duration
	Debug     bool
	Head      bool
	Recording bool
}

func Run(opt Options) ayd.Record {
	timestamp := time.Now()

	logger := &Logger{Status: ayd.StatusHealthy}
	if opt.Debug {
		logger.DebugOut = os.Stderr
	}

	baseDir := os.Getenv("WEBSCENARIO_ARTIFACT_DIR")
	storage, err := NewStorage(baseDir, opt.Target.Opaque, timestamp)
	if err != nil {
		return ayd.Record{
			Time:    timestamp,
			Status:  ayd.StatusFailure,
			Message: err.Error(),
		}
	}

	var browserlog *ayd.Logger
	if opt.Debug {
		f, err := storage.Open("browser.log")
		if err != nil {
			return ayd.Record{
				Time:    timestamp,
				Status:  ayd.StatusFailure,
				Message: err.Error(),
			}
		}
		defer f.Close()
		l := ayd.NewLoggerWithWriter(f, opt.Target)
		browserlog = &l
	}

	ctx, cancel := NewContext(opt.Timeout, opt.Head, browserlog)
	defer cancel()

	env := NewEnvironment(ctx, logger, storage)
	env.EnableRecording = opt.Recording

	stime := time.Now()
	err = env.DoFile(opt.Target.Opaque)
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
