Ayd Web-Scenario plugin
=======================

A headless browser controller for [Ayd](https://github.com/macrat/ayd) status monitoring tool.

- [Quick Start](#quick-start)
- [Reference](reference.md)


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

assert(t("h1").text == "welcome test-id!")
```

Please see also [reference](reference.md) for more information about features you can use in the scenario.

### 3. Test scenario

``` shell
$ ayd-web-scenario --debug /path/to/scenario.lua
```

### 4. Schedule using Ayd

``` shell
$ ayd web-scenario:/path/to/scenario.lua
```
