package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	"github.com/chzyer/readline"
	"github.com/macrat/ayd-web-scenario-plugin/internal"
	"github.com/macrat/ayd/lib-ayd"
	"github.com/spf13/pflag"
)

var (
	Version = "HEAD"
	Commit  = "UNKNOWN"
)

func init() {
	webscenario.Version = Version
	webscenario.Commit = Commit
}

func ParseTargetURL(s string) (mode string, url *ayd.URL, err error) {
	if s == "-" {
		return "repl", &ayd.URL{Scheme: "web-scenario", Opaque: "<stdin>"}, nil
	}

	u, err := ayd.ParseURL(s)
	if err != nil {
		return "", nil, err
	}

	if u.Scheme == "" {
		mode = "standalone"
	} else {
		mode = "ayd"
	}
	u.Scheme = "web-scenario"

	if u.User != nil {
		u.Host = ""
		u.Path = filepath.ToSlash(u.Path)
	} else {
		if u.Opaque == "" {
			u.Opaque = u.Path
			u.Path = ""
		}
		u.Host = ""
		u.Opaque = filepath.ToSlash(u.Opaque)
	}

	return mode, u, nil
}

func main() {
	var arg webscenario.Arg

	flags := pflag.NewFlagSet("ayd-web-scenario-plugin", pflag.ContinueOnError)
	flags.BoolVar(&arg.Debug, "debug", false, "enable debug mode.")
	flags.BoolVar(&arg.Head, "head", false, "show browser window while execution.")
	flags.BoolVar(&arg.Recording, "gif", false, "enable recording animation gif.")
	showVersion := flags.BoolP("version", "v", false, "show version and exit.")
	showHelp := flags.BoolP("help", "h", false, "show help message and exit.")

	cpuprofile := flags.String("cpuprofile", "", "path to cpu profile.")
	flags.MarkHidden("cpuprofile")

	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintf(os.Stderr, "\nPlease see `%s -h` for more information.\n", os.Args[0])
		os.Exit(2)
	}
	switch {
	case *showVersion:
		fmt.Printf("Ayd WebScenaro plugin %s (%s)\n", Version, Commit)
		return
	case *showHelp:
		fmt.Println("Ayd Web-Scenario")
		fmt.Println()
		fmt.Println("Standalone mode:")
		fmt.Println("  $ ayd-web-scenario [OPTIONS] FILE [ARGS...]")
		fmt.Println()
		fmt.Println("REPL mode:")
		fmt.Println("  $ ayd-web-scenario [OPTIONS]")
		fmt.Println("  $ ayd-web-scenario [OPTIONS] - [ARGS...]")
		fmt.Println()
		fmt.Println("Ayd plugin mode:")
		fmt.Println("  $ ayd-web-scenario TARGET_URL")
		fmt.Println()
		fmt.Println("OPTIONS")
		flags.PrintDefaults()
		return
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to create profile file.")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if flags.NArg() == 0 {
		arg.Mode, arg.Target, _ = ParseTargetURL("-")
	} else {
		arg.Args = flags.Args()[1:]

		var err error
		arg.Mode, arg.Target, err = ParseTargetURL(flags.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintf(os.Stderr, "\nPlease see `%s -h` for more information.\n", os.Args[0])
			os.Exit(2)
		}
	}

	switch arg.Mode {
	case "ayd":
		arg.Timeout = 50 * time.Minute
	case "repl":
		if !readline.DefaultIsTerminal() {
			arg.Mode = "stdin"
		}
	}

	rec := webscenario.Run(arg)
	if arg.Mode == "ayd" {
		ayd.NewLogger(arg.Target).Print(rec)
	} else {
		if rec.Status != ayd.StatusHealthy {
			os.Exit(1)
		}
	}
}
