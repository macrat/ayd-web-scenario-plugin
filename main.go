package main

import (
	"fmt"
	"os"
	"time"

	"github.com/macrat/ayd-web-scenario-plugin/internal"
	"github.com/macrat/ayd/lib-ayd"
	"github.com/spf13/pflag"
)

var (
	Version = "HEAD"
	Commit  = "UNKNOWN"
)

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

	rec := webscenario.Run(target, 50*time.Minute, *debugMode, *enableRecording)
	ayd.NewLogger(target).Print(rec)
}
