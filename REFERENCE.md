Reference of Web-Scenario
=========================

The Web-Scenario's script is based on the Lua 5.1 ([GopherLua](https://github.com/yuin/gopher-lua)), and some extra functions.

- [arg](#arg): Read argument and options to this execution.
- [tab](#tab): Open, handle, and close browser tab.
- [element](#element): Read and control HTML element.
- [fetch](#fetch): Communicate via HTTP, without browser.
- [print](#print): Report and store information.
- [assert](#assert): Assertion test values.
- [time](#time): Get or wait time.
- [encoding](#encoding): Serialize or deserialize values.


Arg
---

`arg` is a table includes below properties.

- `arg.target`: The target information that passed via command line argument.
  * `arg.target.url`: The target URL in string. The password in URL will be masked.
  * `arg.target.username`: The username contained in the URL.
  * `arg.target.password()`: A function to get the password contained in the URL.
  * `arg.target.query`: A table contains query values.
  * `arg.target.fragment`: The fragment text in string.

- `arg.alert`: Alerting information. It exists only if it used as alert plugin.
  * `time`: Timestamp in milliseconds.
  * `status`: Status text.
  * `latency`: Probe latency in milliseconds.
  * `target`: Target URL of the probe.
  * `message`: Message string from the probe.
  * `extra`: Extra values the probe reported.

- `arg.mode`: `"ayd"`, `"standalone"`, `"repl"`, or `"stdin"`.

- `arg.debug`: `true` if the `--debug` flag passed.

- `arg.head`: `true` if the `--head` flag passed.

- `arg.recording`: `true` if `--gif` flag passed.


Tab
---

### Create and close tab ###

#### `tab.new([option])`

Make a new tab.

If `option` is nil, the new tab opens `about:blank`.

If `option` is a string, the new tab opens that string as an URL.

If `option` is a table, this function uses below properties.

- `url`: The URL string for the new tab. Default is `about:blank`.
- `width`: The width number of the tab's viewport. Default is 800.
- `height`: The height number of the tab's viewport. Default is 800.
- `recording`: Boolean to enable animated GIF record for the tab. Default is false.

#### `tab:close()`

Close the tab.

#### `tab.viewport`

Get the tab's viewport as a table which has `width` and `height` property.


### Navigate ###

#### `tab:go(url)`

Open the specified `url`.

#### `tab:forward()` / `tab:back()`

Navigate forwards or backwords in the tab's history.

#### `tab:reload()`

Reload the tab.


### Wait and get child elements ###

#### `tab(query)`

Get an [element](#element) using a CSS selector `query`.
This is similar to `document.querySelector` in JavaScript.

This method raise an error if there is no element to match to the `query`.

#### `tab:all(query)`

Get [element](#element)s table using a CSS selector `query`.
This is similar to `document.querySelectorAll` in JavaScript.

The result of this method is a table, but it also can be used as an iterator like below code, rather than normal tables.

``` lua
for elm in tab:all("span") do
  print(elm.text)
end
```

This method never raise an error even if there is no element to match to the `query`.

#### `tab:xpath(query)`

Get [element](#element)s table using a XPath `query`.

The result of this method is a table of [element](#element)s with the same metatable with [`tab:all()`](#taballquery)'s one.

#### `tab:wait(query, [timeout])`

Wait until an element specified in `query` to ready.

It raises an error if the `timeout` in millisecond exceeded.
The default `timeout` is -1 means wait forever.

#### `tab:waitXPath(xpath, [timeout])`

Wait until an element specified in `xpath` to ready.
This function is very similar to [`tab:wait()`](#tabwaitquerytimeout) but it uses XPath instead of CSS selector.


### Retrieve tab information ###

#### `tab.url`

Get current URL the tab is opening.

#### `tab.title`

Get the current page title.

#### `tab:screenshot([name])`

Take a screenshot of current viewport.
If you want a full page screenshot, you can use [`tab("body"):screenshot()`](#elementscreenshotname) instead of this method.

The `name` argument will be used as the file name of screenshot file.
If the `name` omitted, file name will be determined automatically by a serial number.


### Execute JavaScript ###

#### `tab:eval(script)`

Execute JavaScript code in the tab, and returns a value.

``` lua
t:eval([[
  document.querySelector("#something").style.borderColor
]])
```


### Event handling ###

#### `tab:onDialog([callback])`

Set or unset callback function that will called when dialog opened in the tab.
When the callback not set, the browser clicks OK button for all dialogs.

``` lua
t:onDialog(function(dialog)
  print(dialog.type)    -- The type of dialog. "alert", "confirm", "prompt", or "beforeunload".
  print(dialog.message) -- The message on dialog box.
  print(dialog.url)     -- The URL caused this dialog.

  accept = true -- `true` to press OK or YES, or `false` to press CANCEL or something.
  text = "ok!"  -- A string to enter to the "prompt" dialog.

  return accept, text
end)
```

#### `tab:waitDialog([timeout])`

Wait for a dialog shown until `timeout` in millisecond.
It can receive dialogs already shown but not waited yet, unlike [`tab:onDialog()`](#tabondialogcallback).
But this function just receive information and can do nothing to the dialog.

This method returns two values.
The first one is `tab` itself for using method chain.
The second one is information of dialog, that is the same as [`tab:onDialog()`](#tabondialogcallback)'s argument.

#### `tab.dialogs`

Get a dialog list that shown in the tab.
Please see also [`tab:onDialog()`](#tabondialogcallback)

#### `tab:onDownload([callback])`

Set or unset callback function that will called when file downloaded from the tab.

``` lua
t:onDownload(function(file)
  print(file.path)  -- Path to downloaded file in string.
  print(file.bytes) -- The size of downloaded file in bytes.

  return -- Return nothing.
end)
```

#### `tab:waitDownload([timeout])`

Wait for a file downloaded until `timeout` in millisecond.
It can receive downloads already done but not waited yet, unlike [`tab:onDownload()`](#tabondownloadcallback).

This method returns two values.
The first one is `tab` itself for using method chain.
The second one is information of downloaded file, that is the same as [`tab:onDownload()`](#tabdownloadcallback)'s argument.

#### `tab.downloads`

Get a file list that downloaded from the tab.
Please see also [`tab:onDownloaded()`](#tabondownloadcallback)

#### `tab:onRequest(callback)`

Set a callback function that will called when sending network request.

``` lua
t:onRequest(function(req)
  print(req.id)     -- String ID for the request/response.
  print(req.type)   -- The type of the resource. seealso: https://chromedevtools.github.io/devtools-protocol/tot/Network/#type-ResourceType
  print(req.url)    -- The requested URL.
  print(req.method) -- The request method like "GET" or "POST".
  print(req.body)   -- The request body in string. If the request doesn't have post data, it will be a nil value.

  return -- Return nothing.
end)
```

#### `tab:waitRequest([timeout])`

Wait for a network request until `timeout` in millisecond.
It can receive request already done but not waited yet, unlike [`tab:onRequest()`](#tabonrequestcallback).

This method returns two values.
The first one is `tab` itself for using method chain.
The second one is information of the request, that is the same as [`tab:onRequest()`](#tabonrequestcallback)'s argument.

#### `tab.requests`

Get a request list that sent from the tab.
Please see also [`tab:onRequest()`](#tabonrequestcallback)

#### `tab:onResponse(callback)`

Set a callback function that will called when network responce received.

``` lua
t:onResponse(function(res)
  print(res.id)         -- String ID for the request/response.
  print(res.type)       -- The type of the resource. seealso: https://chromedevtools.github.io/devtools-protocol/tot/Network/#type-ResourceType
  print(res.url)        -- The requested URL.
  print(res.status)     -- The status code of the response like 200 or 404.
  print(res.mimetype)   -- The MIME-Type of the response body.
  print(res.remoteIP)   -- The server's IP address.
  print(res.remotePort) -- The server's network port.
  print(res.length)     -- The received body's length transported over network. This is not actual size of the body if the response compressed or encoded.

  -- The response table have `read` and `lines` to read response body. The usage of these functions are the same as Lua's file object.
  -- If body is too large, these methods raise error.
  print(res:read("all"))

  return -- Return nothing.
end)
```

#### `tab:waitResponse([timeout])`

Wait for a network response received until `timeout` in millisecond.
It can receive response already done but not waited yet, unlike [`tab:onResponse()`](#tabonresponsecallback).

This method returns two values.
The first one is `tab` itself for using method chain.
The second one is information of the response, that is the same as [`tab:onResponse()`](#tabonresponsecallback)'s argument.

#### `tab.responses`

Get a response list that the tab received.
Please see also [`tab:onResponse()`](#tabonresponsecallback)


Element
-------

### Retrieve information ###

#### `element.text` 

Get inner text of the element.

#### `element.innerHTML`

Get inner HTML of the element, as a string.

#### `element.outerHTML`

Get outer HTML of the element, as a string.

#### `element.value`

Get *value* of the element.
This property can be used for HTML elements have `.value` property in JavaScript, like **input**.

#### `element[property]`

Get element's HTML property by name.

``` lua
t("a")["href"] -- Get the URL of A tag.
```

#### `element:screenshot([name])`

Take a screenshot of the element.
If you want a viewport screenshot, you can use [`tab:screenshot()`](#tabscreenshotname) instead of this method.

The `name` argument will be used as the file name of screenshot file.
If the `name` omitted, file name will be determined automatically by a serial number.


### Input and control ###

#### `element:sendKeys(keys, [modifiers])`

Send `keys` into the element.
You can use a global variable named `key` to send special keys such as `key.backspace` or `key.f1`.

The `modifiers` accepts a list of key modifiers, such as `"alt"`, `"ctrl"`, `"meta"`, and `"shift"`.

``` lua
-- send "helle", backspace, and "o world" to the input.
t("input[type=text]"):sendKeys("helle" .. key.backspace .. "o world")

-- send Ctrl-A and Ctrl-C to the textarea.
t("textarea"):sendKeys("ac", {"ctrl"})
```

#### `element:setValue(value)`

Set `value` into the value of element.
This method can be used for HTML elements have `.value` property in JavaScript, like **input**.

#### `element:click([button])`

Click on the element.

The `button` is a button name, `"left"`, `"middle"`, `"right"`, `"back"`, or `"forward"`. If omit this, it clicks left mouse button.

#### `element:submit()`

Submit the form contains the element.

#### `element:focus()` / `element:blur()`

Set or unset focus on the element.


### Get child elements ###

#### `element(query)`

Get an [element](#element) by `query` from children of this element.
Please see also [`tab(query)`](#tabquery).

#### `element:all(query)`

Get [element](#element)s table from children of this element.
Please see also [`tab:all(query)`](#taballquery).


Fetch
-----

#### `fetch(url, [options])`

Send HTTP/HTTPS request to `url`, and wait for response.

The `options` is a table and can have below fields.

- `method`: HTTP method in string such as `"GET"` or `"POST"`. The default is `"GET"` normally, but it is `"POST"` if set non-nil value to `body`.
- `headers`: A table that contains header key-values.
- `body`: The body value for POST or PUT method. It is a string, a number, or an iterator function that returns each lines in string.
- `timeout`: Timeout duration in millisecond. The default is 5 minutes.

The first return value is a table that response from the server, contains below fields.

- `url`: URL of this resource.
- `status`: HTTP response status like `200` for OK.
- `headers`: HTTP headers server sent.
- `length`: The transfered length in bytes.
- `read`: A method for read the response body. This is the same usage as [`file:read`](https://www.lua.org/manual/5.1/manual.html#pdf-file:read)
- `lines`: A method to make an iterator function to read body.
- `cookiejar`: Cookie store to continue session from previous fetch.

The second return value is a cookie jar that holds all cookies set while the fetch.
You can read cookies for specific URL using `get(url)` method, or all cookies using `all()` method.


Print
-----

#### `print(values...)`

Print values as message output of the execution.

The output will be buffered and doesn't print anything immediately, in the default mode.
If you want to see them in real time, you can use `--debug` flag when execute `ayd-web-scenario` command.

#### `print.extra(key, value)`

Set an extra value for the execution result.
The extra values will be stored in the Ayd execution log as is as possible.

#### `print.status(status)`

Set the status of the execution.
The `status` argument must be one of `"healthy"`, `"unknown"`, `"degrade"`, `"failure"`, or `"aborted"`.


Assert
------

#### `assert(test, [message])`

Raises error if `test` is false or nil value. Otherwise, it returns all arguments.
This is the same function as [the original function](https://www.lua.org/manual/5.1/manual.html#pdf-assert) of Lua 5.1.

#### `assert.eq(x, y)`

Raises error if `x` and `y` are not the same.
This is similar to [`assert(x == y)`](#asserttestmessage), but it provides more convinient error message.

#### `assert.ne(x, y)`

Raises error if `x` and `y` are the same.
This is similar to [`assert(x ~= y)`](#asserttestmessage), but it provides more convinient error message.

#### `assert.lt(x, y)`

Raises error if `x` is not smaller than `y`.
This is similar to [`assert(x < y)`](#asserttestmessage), but it provides more convinient error message.

#### `assert.le(x, y)`

Raises error if `x` is not smaller or equals to `y`.
This is similar to [`assert(x <= y)`](#asserttestmessage), but it provides more convinient error message.

#### `assert.gt(x, y)`

Raises error if `x` is not greater than `y`.
This is similar to [`assert(x > y)`](#asserttestmessage), but it provides more convinient error message.

#### `assert.ge(x, y)`

Raises error if `x` is not greater or equals to `y`.
This is similar to [`assert(x >= y)`](#asserttestmessage), but it provides more convinient error message.


Time
----

#### `time.now()`

Get current time in UNIX time, milliseconds.

#### `time.sleep(milliseconds)`

Pause program for specified `milliseconds`.

It's highly recommended to use [`tab:wait()`](#tabwaitquerytimeout) method if possible, for execution speed and stability reasons.

#### `time.millisecond` / `time.second` / `time.minute` / `time.hour` / `time.day` / `time.year`

Helper numbers to build a time in milliseconds.

``` lua
time.sleep(0.5*time.hour + 10*time.second)  -- Wait for a half hour and 10 seconds.
```

#### `time.format(millisecond, [format])`

Convert a `millisecond` number to a human readable string.
The default `format` is `"%Y-%m-%dT%H:%M:%S%z"` aka RFC3339 format.


Encoding
--------

### JSON ###

#### `fromjson(json)`

Parse `json` string.

#### `tojson(value)`

Encode `value` into JSON string.


### CSV ###

#### `fromcsv(csv, [useheader])`

Parse `csv` and make an iterator function.
The parameter `csv` can be a string, a list table, or an iterator function that returns strings.

In default, the first row is used as the header, and each results will be key-value style.
If the parameter `useheader` is `false`, this function does not use the first row as header and each results will be a list style.

The first of the result is an iterator function that returns each row values.
And the second of the result is a list of header.

``` lua
iter, header = fromcsv(io.open("path/to/input.csv"):lines())
print(header)
for row in iter do
  print(row)
end
```

#### `tocsv(values, [header])`

Make an iterator that encode `values` to CSV lines.

The parameter `values` can be a table of tables, or an iterator function that returns a table.

If the parameter `header` is a table that list of string, the table will be used as the header.

If `header` is absent or `true`, the keys of the first element will be used as the header.
The columns is sorted by header name in this mode.

If `header` is `false`, this function does not care about header.
Only values indexed by number will be included in the result, in this case.

If the `header` is not `false`, you can use values indexed by string or number, in each row' table.
Values that matched name by string has priority.

``` lua
iter = tocsv({
  {hello="world", foo="bar"},
  {"number", "indexed"},
  {"you can", foo="combine"}
}, {"hello", "foo"})

f = io.open("path/to/output.csv")
for row in iter do
  f:write(row .. "\n")
end
f:close()
```

``` csv
hello,foo
world,bar
number,indexed
you can,combine
```


### XML ###

#### `fromxml(xml)`

Parse `xml` as a XML and returns a table.
The parameter `xml` can be a string, a list table, or an iterator function that returns strings.

``` lua
assert.eq(
  fromxml([[
    <feed xmlns="http://www.w3.org/2005/Atom">
      <title>foobar</title>
      <link href="https://example.com" />
    </feed>
  ]]),

  {"feed", xmlns="http://www.w3.org/2005/Atom"
    {"title",
      "foobar"
    },
    {"link", href="https://example.com"}
  }
)
```

#### `toxml(table)`

Encode a `table` as a XML string.

The `table[1]` is used as a tag name.
Second or later elements is used as children of the tag.
Other properties named by string is used as attributes.

Please see also the example for [`fromxml`](#fromxmlxml).
