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

func NewContext(debug bool) (context.Context, context.CancelFunc) {
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
	if !debug {
		opts = append(opts, chromedp.Headless)
	}

	ctx, cancel1 := context.WithTimeout(context.Background(), time.Hour)
	ctx, cancel2 := chromedp.NewExecAllocator(ctx, opts...)
	ctx, cancel3 := StartBrowser(ctx)

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
	logger := &Logger{Debug: debug, Status: ayd.StatusHealthy}

	ctx, cancel := NewContext(debug)
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
