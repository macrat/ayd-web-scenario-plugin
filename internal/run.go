package webscenario

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/macrat/ayd/lib-ayd"
)

func getenv(name ...string) string {
	for _, n := range name {
		if v := os.Getenv(n); v != "" {
			return v
		}
	}
	return ""
}

func NewExecAllocator(ctx context.Context, withHead bool) (context.Context, context.CancelFunc) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.WindowSize(800, 800),
		chromedp.ProxyServer(getenv("webscenario_proxy", "WEBSCENARIO_PROXY", "all_proxy", "ALL_PROXY", "https_proxy", "HTTPS_PROXY", "http_proxy", "HTTP_PROXY")),

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

func NewContext(arg Arg, debuglog *ayd.Logger) (context.Context, context.CancelFunc) {
	ctx := context.Background()

	stopTimeout := func() {}
	if arg.Timeout > 0 {
		ctx, stopTimeout = context.WithTimeout(ctx, arg.Timeout)
	}

	stopNotify := func() {}
	if arg.Mode != "repl" {
		ctx, stopNotify = signal.NotifyContext(ctx, os.Interrupt)
	}

	ctx, stopAllocator := NewExecAllocator(ctx, arg.Head)

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

func Run(arg Arg) ayd.Record {
	timestamp := time.Now()

	logger := &Logger{Status: ayd.StatusHealthy, Debug: arg.Debug}
	if arg.Mode != "ayd" {
		logger.Stream = os.Stdout
	}

	baseDir := os.Getenv("WEBSCENARIO_ARTIFACT_DIR")
	storage, err := NewStorage(arg.ArtifactDir(baseDir), timestamp)
	if err != nil {
		return ayd.Record{
			Time:    timestamp,
			Status:  ayd.StatusFailure,
			Message: err.Error(),
		}
	}

	var browserlog *ayd.Logger
	if arg.Debug {
		f, err := storage.Open("browser.log")
		if err != nil {
			return ayd.Record{
				Time:    timestamp,
				Status:  ayd.StatusFailure,
				Message: err.Error(),
			}
		}
		defer f.Close()
		l := ayd.NewLoggerWithWriter(f, arg.Target)
		browserlog = &l
	}

	ctx, cancel := NewContext(arg, browserlog)
	defer cancel()

	env := NewEnvironment(ctx, logger, storage, arg)
	env.EnableRecording = arg.Recording

	var latency time.Duration
	switch arg.Mode {
	case "repl":
		err = env.DoREPL(ctx)
	case "stdin":
		err = env.DoStream(os.Stdin, "<stdin>")
	default:
		stime := time.Now()
		err = env.DoFile(arg.Path())
		latency = time.Since(stime)
	}

	env.Close()
	logger.HandleError(ctx, err)

	if xs := storage.Artifacts(); len(xs) > 0 {
		logger.SetExtra("artifacts", xs)
	}

	return logger.AsRecord(timestamp, latency)
}
