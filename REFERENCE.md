Reference of Web-Scenario
=========================

The Web-Scenario's script is based on the Lua 5.1 ([GopherLua](https://github.com/yuin/gopher-lua)), and some extra functions.

- [tab](#tab): Open, handle, and close browser tab.
- [element](#element): Read and control HTML element.
- [print](#print): Report and store information.
- [time](#time): Get or wait time.


Tab
---

### Create and close tab ###

#### `tab.new([url])`

Make a new tab. It opens `about:blank` if `url` is not specified.

#### `tab:close()`

Close the tab.

#### `tab.viewport` / `tab:setViewport(width, height)`

Get or set the tab's viewport.


### Navigate ###

#### `tab:go(url)`

Open the specified `url`.

#### `tab:forward()` / `tab:back()`

Navigate forwards or backwords in the tab's history.

#### `tab:reload()`

Reload the tab.


### Wait and get child elements ###

#### `tab(query)`

Get an [element](#element) using a `query`.
This is similar to `document.querySelector` in JavaScript.

This method raise an error if there is no element to match to the `query`.

#### `tab:all(query)`

Get [element](#element)s table using a `query`.
This is similar to `document.querySelectorAll` in JavaScript.

The result of this method is a table, but it also can be used as an iterator like below code, rather than normal tables.

``` lua
for elm in tab:all("span") do
  print(elm.text)
end
```

This method never raise an error even if there is no element to match to the `query`.

#### `tab:wait(query)`

Wait until an element specified in `query` to ready.


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

#### `tab:recording(bool)`

Enable or disable animated GIF recording.


### Execute JavaScript ###

#### `tab:eval(script)`

Execute JavaScript code in the tab, and returns a value.

``` lua
t:eval([[
  document.querySelector("#something").style.borderColor
]])
```


### Event handling ###

You can set only one callback function for each events.
The previous callback function will be disabled when you set a new callback function.

#### `tab:onDialog(callback)`

Set a callback function that will called when dialog opened in the tab.
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

#### `tab:onDownloaded(callback)`

Set a callback function that will called when file downloaded from the tab.

``` lua
t:onDownloaded(function(file)
  print(file.path)  -- Path to downloaded file in string.
  print(file.bytes) -- The size of downloaded file in bytes.

  return -- Return nothing.
end)
```

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

  -- Read the body of the respnse using body function.
  -- This method returns a string if succeeded to get body. Otherwise, for instance if body is too large, it returns nil.
  print(res:body())

  return -- Return nothing.
end)
```


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

#### `element:sendKeys(keys)`

Send `keys` into the element.

#### `element:setValue(value)`

Set `value` into the value of element.
This method can be used for HTML elements have `.value` property in JavaScript, like **input**.

#### `element:click()`

Click on the element.

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


Time
----

#### `time.now()`

Get current time in UNIX time, milliseconds.

#### `time.sleep(milliseconds)`

Pause program for specified `milliseconds`.

It's highly recommended to use [`tab:wait()`](#tabwaitquery) method if possible, for execution speed and stability reasons.

#### `time.millisecond` / `time.second` / `time.minute` / `time.hour`

Helper numbers to build a time in milliseconds.

``` lua
time.sleep(0.5*time.hour + 10*time.second)  -- Wait for a half hour and 10 seconds.
```

#### `time.format(millisecond)`

Convert a `millisecond` number to `YYYY-mm-dd"T"HH:MM:ssZ` formatted string aka RFC3339 format.
