Ayd Web-Scenario plugin
=======================

A headless browser controller for [Ayd](https://github.com/macrat/ayd) status monitoring tool.

- [Quick Start](#quick-start)
- [Reference](REFERENCE.md)


## Quick Start

### 1. Install

~~Download a plugin binary from [release page](https://github.com/macrat/ayd-web-scenario-plugin/releases).~~ (pre built binary is not yet released)
And place binary to some directory that is included in PATH environment variable.

### 2. Make a scenario

A scenario is a script to control headless browser, written in [Lua](https://www.lua.org/).

A scenario looks lie above.

``` lua
t = tab.new("https://your-service.example.com")

t("input[name=username]"):sendKeys("test-id")
t("input[name=password]"):sendKeys("test-password")
t("input[type=submit]"):click()

assert.eq(t("h1").text, "welcome test-id!")
```

Please see also [reference](reference.md) for more information about features you can use in the scenario.

You can use REPL mode for testing how to write scenario. Please execute `ayd-web-scenario` without any argument.

### 3. Test scenario in standalone mode

If you passed file path instead of URL, web-scenario works in the standalone mode that shows logs more readable style.
You can use `--head` flag for check what is going on on the window, and/or `--debug` flag for get more detail information.

``` shell
$ ayd-web-scenario /path/to/scenario.lua
```

### 4. Schedule using Ayd

You can use Web-Scenario as a plugin of Ayd for monitoring web services.

``` shell
$ ayd web-scenario:/path/to/scenario.lua
```
