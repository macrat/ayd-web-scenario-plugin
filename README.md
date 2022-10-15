Ayd WebScenario plugin
======================

A headless browser controller for [Ayd](https://github.com/macrat/ayd) status monitoring tool.


## Quick start

### 1. Install

~~Download a plugin binary from [release page](https://github.com/macrat/ayd-web-scenario-plugin/releases).~~ (pre build binary is not yet released)
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

### 3. Run Ayd

``` shell
$ web-scenario:/path/to/scenario.lua
```


## Reference

The web scinario script is based on lua 5.1 ([GopherLua](https://github.com/yuin/gopher-lua)), and some extra functions.

### Tab

#### create and close tab

- `tab.new([url])`

  Make a new tab. It opens `about:blank` if `url` is not specified.

- `tab:close()`

  Close the tab.

#### navigate

- `tab:go(url)`

  Open the specified `url`.

- `tab:forward()` / `tab:back()`

  Navigate forwards or backwords in the tab's history.

- `tab:reload()`

  Reload the tab.

#### wait and get child elements

- `tab(query)` / `tab:all(query)`

  Get element(s) using a query. These are similar to `document.querySelector` or `document.querySelectorAll`.

- `tab:wait(query)`

  Wait until an element specified in query to ready.

### retrieve tab information

- `tab.url`

  Get current URL the tab is opening.

- `tab.title`

  Get the current page title.

- `tab:screenshot([name])`

  Take a screenshot of current viewport.
  If you want a full page screenshot, you can use `tab("body"):screenshot()` instead of this method.

- `tab:eval(script)`

  Execute JavaScript code in the tab, and returns a value.

  e.g. `t:eval([[ document.querySelector("#something").style.borderColor ]])`

### settings

- `tab.viewport` / `tab:setViewport(width, height)`

  Get or set the tab's viewport.

- `tab:onDialog(callback)`

  Set a callback function that will called when dialog opened in the tab.

  The callback function has three arguments, `type`, `message`, and `url`.
  * `type` means dialog type, one of `"alert"`, `"confirm"`, `"prompt"`, and `"beforeunload"`.
  * `message` is the message on the dialog box.
  * `url` is the URL caused this dialog.

  You can return two values from the callback function, `accept` and `text`.
  * If `accept` is true or absent, it will click on `"OK"`. Otherwise it will click on `"cancel"` or something.
  * `text` value will used for the prompt input value, if the dialog type was `"prompt"`.

  __NOTE__: It disables the previous callback function when called.

- `tab:onDownloaded(callback)`

  Set a callback function that will called when file downloaded from the tab.

  The callback function has two arguments, `filepath` and `bytes`.
  * `filepath` is the path to downloaded file.
  * `bytes` is downloaded file size in bytes.

  __NOTE__: It disables the previous callback function when called.


## element / elementsarray

### retrieve information

- `element.text` / `element.innerHTML` / `element.outerHTML` / `element.value`

  Get text, innerHTML, outerHTML, value of the element.
  These are the same as JavaScript's property.

- `elementsarray.text` / `elementsarray.innerHTML` / `elementsarray.outerHTML` / `elementsarray.value`

  Get text, innerHTML, outerHTML, value of the elements.
  There are almost the same as `element`'s one, but returns an array of strings.

- `element[property]`

  Get element's HTML property by name.

  e.g. `elm["href"]`

- `element:screenshot([name])`

  Take a screenshot of the element.

  __NOTE__: `elementsarray` doesn't have these methods.

### input and control

- `element:sendKeys(keys)` / `elementsarray:sendKeys(keys)`

  Send keys into the element(s).

- `element:setValue(value)` / `elementsarray:setValue(value)`

  Set value into the element(s).
  These methods can used for HTML elements have `.value` property in JavaScript.

- `element:click()` / `elementsarray:click()`

  Click on the element(s).

- `element:submit()`

  Submit the form contains the element.

  __NOTE__: `elementsarray` doesn't have this method.

- `element:focus()` / `element:blur()`

  Set or unset focus on the element.

  __NOTE__: `elementsarray` doesn't have these methods.

### get child elements

- `element(query)` / `elementsarray(query)`

  Get an element from children of this element(s).
  `element(query)` returns a single `element`, but `elementsarray(query)` returns `elementsarray`.

- `element:all(query)` / `elementsarray:all(query)`

  Get elements from children of this element(s).
  Both methods returns `elementsarray`.
