package webscenario

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/macrat/ayd-web-scenario/internal/lua"
	"github.com/macrat/ayd/lib-ayd"
)

func init() {
	os.Setenv("TZ", "UTC")
	os.Setenv("PATH", "./testdata/windows-compat"+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func StartTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			target = "world"
		}
		fmt.Fprintf(w, `<title>%s - test</title><div id="greeting">hello <b class="target">%s</b>!</div>`, target, target)
	})

	mux.HandleFunc("/complex-dom", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, `
			<div><h1>text</h1><b>hello </b>beautiful <b class="target">world</b><span>!</span></div>
			<form action=GET><h1>form</h1><input type="text"><input type="text"></form>
		`)
	})

	count := 0
	mux.HandleFunc("/counter", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		count++
		fmt.Fprintf(w, `current count is <span>%d</span>`, count)
	})

	mux.HandleFunc("/dynamic", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, `
			<div>
				<ol></ol>
				<script>count = 0</script>
				<button id=append onclick="document.querySelector('ol').innerHTML += '<li>count=' + count + '</li>'; count++;">append</button>
			</div>

			<div>
				<span id="text"></span>
				<input type=text onchange="document.querySelector('#text').innerText = event.target.value">
			</div>

			<div>
				<span id=look-at-me onfocus="event.target.innerText = 'focus'" onblur="event.target.innerText = 'blur'" tabindex=-1>blur</span>
			</div>

			<form>
				<div id=submitted>%s</div>
				<textarea name="textarea"></textarea>
				<input type=submit>
			</form>

			<div>
				<div id="key-listener" tabIndex=-1></div>
				<script>const kl = document.querySelector('#key-listener'); kl.addEventListener('keydown', (ev) => kl.innerText = (ev.altKey ? 'alt+' : '') + (ev.ctrlKey ? 'ctrl+' : '') + (ev.metaKey ? 'meta+' : '') + (ev.shiftKey ? 'shift+' : '') + ev.key)</script>
			</div>

			<div>
				<div id="click-listener" tabIndex=-1 style="height: 1em"></div>
				<script>const cl = document.querySelector('#click-listener'); cl.addEventListener('mousedown', (ev) => cl.innerText = ev.button)</script>
			</div>
		`, r.URL.Query().Get("textarea"))
	})

	mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		if r.Method == "POST" {
			fmt.Fprintf(w, `
				<span>%s</span>
			`, r.FormValue("value"))
		} else {
			fmt.Fprintf(w, `
				<form method=POST>
					<input name=value /><input type=submit />
				</form>
			`)
		}
	})

	mux.HandleFunc("/window-size", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, `
			<span>loading...</span>
			<script>
				function onResize() {
					document.querySelector('span').innerText = window.innerWidth + 'x' +window.innerHeight;
				}
				window.addEventListener('resize', onResize);
				onResize();
			</script>
		`)
	})

	mux.HandleFunc("/dialog/alert", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, `<script>alert('welcome!')</script>`)
	})
	mux.HandleFunc("/dialog/confirm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, `<span></span><script>document.querySelector('span').innerText = confirm('are you sure?')</script>`)
	})
	mux.HandleFunc("/dialog/prompt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, `<span></span><script>document.querySelector('span').innerText = JSON.stringify(prompt('type something here!'))</script>`)
	})

	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<a href="/download/data.txt">download</a>`)
	})
	mux.HandleFunc("/download/data.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/octet-stream")
		fmt.Fprintf(w, `this is a data`)
	})

	mux.HandleFunc("/header", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s %q", r.Method, r.Header.Get("X-Header-Test"))
	})
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, r.Body)
	})
	mux.HandleFunc("/cookie/set", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "cookie_test",
			Value: "hello world",
		})
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/cookie/get", func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("cookie_test")
		if err != nil {
			fmt.Fprint(w, "not set")
		} else {
			fmt.Fprint(w, c.Value)
		}
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "something wrong!")
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprint(w, "ok")
	})

	return httptest.NewServer(mux)
}

func RegisterTestUtil(L *lua.State, storage *Storage, server *httptest.Server) {
	L.CreateTable(0, 2)
	L.SetFuncs(-1, map[string]lua.GFunction{
		"url": func(L *lua.State) int {
			if L.Type(1) == lua.String {
				L.PushString(server.URL + L.ToString(1))
			} else {
				L.PushString(server.URL)
			}
			return 1
		},
		"storage": func(L *lua.State) int {
			if L.Type(1) == lua.String {
				L.PushString(filepath.Join(storage.Dir, L.ToString(1)))
			} else {
				L.PushString(storage.Dir)
			}
			return 1
		},
	})
	L.SetGlobal("TEST")
}

type DebugWriter testing.T

func (w *DebugWriter) Write(b []byte) (int, error) {
	(*testing.T)(w).Log(strings.TrimSuffix(string(b), "\n"))
	return len(b), nil
}

func Test_testSenarios(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/scenario/*.lua")
	if err != nil {
		t.Fatalf("failed to get tests: %s", err)
	}

	server := StartTestServer()
	t.Cleanup(server.Close)

	ctx, cancel := NewContext(Arg{Mode: "ayd", Timeout: 5 * time.Minute}, nil)
	t.Cleanup(cancel)

	target, _ := ayd.ParseURL("web-scenario://foo:bar@/dummy/script.lua?hello=world&hoge=fuga#piyo")

	for _, p := range files {
		p := p
		b := filepath.Base(p)
		if strings.HasPrefix(b, "_") {
			continue
		}
		t.Run(b, func(t *testing.T) {
			t.Parallel()

			s, err := NewStorage(t.TempDir(), time.Now())
			if err != nil {
				t.Fatalf("failed to prepare storage: %s", err)
			}

			logger := &Logger{Stream: (*DebugWriter)(t)}
			env, err := NewEnvironment(ctx, logger, s, Arg{Mode: "ayd", Args: []string{"abc", "def"}, Target: target})
			if err != nil {
				t.Fatalf("failed to prepare environment: %s", err)
			}
			defer env.Close()

			RegisterTestUtil(env.lua, s, server)

			if err := env.DoFile(p); err != nil {
				t.Fatalf(err.Error())
			}
		})
	}
}

func Test_errorInEvent(t *testing.T) {
	t.Parallel()

	server := StartTestServer()
	t.Cleanup(server.Close)

	ctx, cancel := NewContext(Arg{Mode: "ayd", Timeout: 5 * time.Minute}, nil)
	t.Cleanup(cancel)

	target, _ := ayd.ParseURL("web-scenario://foo:bar@/dummy/script.lua?hello=world&hoge=fuga#piyo")

	s, err := NewStorage(t.TempDir(), time.Now())
	if err != nil {
		t.Fatalf("failed to prepare storage: %s", err)
	}

	logger := &Logger{Stream: (*DebugWriter)(t)}
	env, err := NewEnvironment(ctx, logger, s, Arg{Mode: "ayd", Args: []string{"abc", "def"}, Target: target})
	if err != nil {
		t.Fatalf("failed to prepare environment: %s", err)
	}
	defer env.Close()

	RegisterTestUtil(env.lua, s, server)

	expect := `testdata/error-in-event.lua:4: test error
stack traceback:
	[G]: in function 'error'
	testdata/error-in-event.lua:4: in main chunk
	[G]: ?`

	if err := env.DoFile("testdata/error-in-event.lua"); err == nil {
		t.Fatalf("expected error but got nil")
	} else if err.Error() != expect {
		t.Fatalf("unexpected error:\n%s", err)
	}
}
