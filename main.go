package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/macrat/ayd/lib-ayd"
	"github.com/spf13/pflag"
	"github.com/yuin/gopher-lua"
)

var (
	Version = "HEAD"
	Commit  = "UNKNOWN"
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

func NewContext(timeout time.Duration, debuglog *ayd.Logger) (context.Context, context.CancelFunc) {
	ctx, stopTimeout := context.WithTimeout(context.Background(), timeout)
	ctx, stopNotify := signal.NotifyContext(ctx, os.Interrupt)
	ctx, stopAllocator := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)
	ctx, stopBrowser := StartBrowser(ctx, debuglog)

	return ctx, func() {
		stopBrowser()
		stopAllocator()
		stopNotify()
		stopTimeout()
	}
}

func RunWebScenario(target *ayd.URL, timeout time.Duration, debug bool, enableRecording bool, callback func(ayd.Record)) {
	timestamp := time.Now()

	logger := &Logger{Debug: debug, Status: ayd.StatusHealthy}

	baseDir := os.Getenv("WEBSCENARIO_ARTIFACT_DIR")
	storage, err := NewStorage(baseDir, target.Opaque, timestamp)
	if err != nil {
		callback(ayd.Record{
			Time:    timestamp,
			Status:  ayd.StatusFailure,
			Message: err.Error(),
		})
		return
	}

	var browserlog *ayd.Logger
	if debug {
		f, err := storage.Open("browser.log")
		if err != nil {
			callback(ayd.Record{
				Time:    timestamp,
				Status:  ayd.StatusFailure,
				Message: err.Error(),
			})
			return
		}
		defer f.Close()
		l := ayd.NewLoggerWithWriter(f, target)
		browserlog = &l
	}

	ctx, cancel := NewContext(timeout, browserlog)
	defer cancel()

	env := NewEnvironment(ctx, logger, storage)
	env.EnableRecording = enableRecording
	defer env.Close()

	stime := time.Now()
	err = env.DoFile(target.Opaque)
	latency := time.Since(stime)

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
	callback(r)
}

func ParseTargetURL(s string) (*ayd.URL, error) {
	u, err := ayd.ParseURL(s)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "web-scenario"
	}
	if u.Opaque == "" {
		u.Opaque = u.Path
		u.Path = ""
	}
	u.Host = ""
	return u, nil
}

func main() {
	flags := pflag.NewFlagSet("ayd-web-scenario-plugin", pflag.ContinueOnError)
	debugMode := flags.Bool("debug", false, "enable debug mode.")
	enableRecording := flags.Bool("gif", false, "enable recording animation gif.")
	showVersion := flags.BoolP("version", "v", false, "show version and exit.")
	showHelp := flags.BoolP("help", "h", false, "show help message and exit.")

	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintf(os.Stderr, "\nPlease see `%s -h` for more information.\n", os.Args[0])
		os.Exit(2)
	}
	switch {
	case *showVersion:
		fmt.Printf("Ayd WebScenaro plugin %s (%s)\n", Version, Commit)
		return
	case *showHelp || len(flags.Args()) != 1:
		fmt.Println("$ ayd-web-scenario-plugin [OPTIONS] TARGET_URL\n\nOptions:")
		flags.PrintDefaults()
		return
	}

	target, err := ParseTargetURL(flags.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintf(os.Stderr, "\nPlease see `%s -h` for more information.\n", os.Args[0])
		os.Exit(2)
	}

	RunWebScenario(target, 50*time.Minute, *debugMode, *enableRecording, func(r ayd.Record) {
		ayd.NewLogger(target).Print(r)
	})
}
