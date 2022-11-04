package webscenario

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

func isIncomplete(err error) bool {
	if lerr, ok := err.(*lua.ApiError); ok {
		if perr, ok := lerr.Cause.(*parse.Error); ok {
			return perr.Pos.Line == parse.EOF
		}
	}
	return false
}

func (env *Environment) DoREPL(ctx context.Context) error {
	rl, err := readline.New("> ")
	if err != nil {
		return err
	}
	defer rl.Close()

	fmt.Fprintf(rl, "Ayd Web-Scenario %s (%s)\n", Version, Commit)

	var code string

	for {
		env.Unlock()

		if code == "" {
			rl.SetPrompt("> ")
		} else {
			rl.SetPrompt(">> ")
		}

		line, err := rl.Readline()
		if err == io.EOF {
			env.Lock()
			return nil
		} else if err == readline.ErrInterrupt {
			fmt.Fprintln(rl, "keyboard interrupt")
			code = ""
			rl.Clean()
			env.Lock()
			continue
		} else if err != nil {
			env.Lock()
			return err
		}
		if code == "" {
			code = line
		} else {
			code = code + "\n" + line
		}

		env.Lock()
		if f, err := env.lua.Load(strings.NewReader("return "+code), "<stdin>"); err == nil {
			env.lua.Push(f)
		} else if f, err := env.lua.Load(strings.NewReader(code), "<stdin>"); err == nil {
			env.lua.Push(f)
		} else if isIncomplete(err) {
			continue
		} else {
			env.logger.HandleError(ctx, err)
			code = ""
			continue
		}

		if err = env.lua.PCall(0, lua.MultRet, nil); err != nil {
			env.logger.HandleError(ctx, err)
		}

		if (code == "exit" || code == "quit" || code == "bye") && env.lua.GetTop() == 1 && env.lua.Get(1).Type() == lua.LTNil {
			fmt.Fprintln(rl, "Use os.exit() or Ctrl-D to exit.")
		} else {
			var xs []string
			for i := 1; i <= env.lua.GetTop(); i++ {
				xs = append(xs, string(env.lua.ToStringMeta(env.lua.Get(i)).(lua.LString)))
			}
			if len(xs) > 0 {
				fmt.Fprintln(rl, strings.Join(xs, "\t"))
			}
		}

		env.lua.Pop(env.lua.GetTop())
		code = ""
	}
}

func (env *Environment) DoStream(r io.Reader, name string) error {
	f, err := env.lua.Load(r, name)
	if err != nil {
		return err
	}
	env.lua.Push(f)
	return env.lua.PCall(0, 0, nil)
}
