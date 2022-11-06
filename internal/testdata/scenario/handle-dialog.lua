t = tab.new()
assert.eq(t.dialogs, {
    _waited=0,
})

t:go(TEST.url("/dialog/alert"))

t2 = tab.new(TEST.url())
t2:onDialog(function(dialog)
    error(string.format("caught unexpected dialog: %s", tojson(dialog)))
end)


assert.eq(t.dialogs, {
    {url=TEST.url("/dialog/alert"), type="alert", message="welcome!"},
    _waited=0,
})
_, dialog = t:waitDialog()
assert.eq(dialog, {url=TEST.url("/dialog/alert"), type="alert", message="welcome!"})
assert.eq(t.dialogs, {
    {url=TEST.url("/dialog/alert"), type="alert", message="welcome!"},
    _waited=1,
})


called = false
t:onDialog(function(dialog)
    called = true
    assert.eq(dialog, {
        type    = "alert",
        message = "welcome!",
        url     = TEST.url("/dialog/alert"),
    })
    return true
end)
assert.eq(called, false)
t:go(TEST.url("/dialog/alert"))
assert.eq(called, true)


assert.eq(t.dialogs, {
    {url=TEST.url("/dialog/alert"), type="alert", message="welcome!"},
    {url=TEST.url("/dialog/alert"), type="alert", message="welcome!"},
    _waited=1,
})


called = false
t:onDialog(function(dialog)
    called = true
    assert.eq(dialog, {
        type    = "confirm",
        message = "are you sure?",
        url     = TEST.url("/dialog/confirm"),
    })
    return true
end)
assert.eq(called, false)
t:go(TEST.url("/dialog/confirm"))
assert.eq(called, true)
assert.eq(t("span").text, "true")


assert.eq(t.dialogs, {
    {url=TEST.url("/dialog/alert"), type="alert", message="welcome!"},
    {url=TEST.url("/dialog/alert"), type="alert", message="welcome!"},
    {url=TEST.url("/dialog/confirm"), type="confirm", message="are you sure?"},
    _waited=1,
})
t:waitDialog():waitDialog()
assert.eq(t.dialogs, {
    {url=TEST.url("/dialog/alert"), type="alert", message="welcome!"},
    {url=TEST.url("/dialog/alert"), type="alert", message="welcome!"},
    {url=TEST.url("/dialog/confirm"), type="confirm", message="are you sure?"},
    _waited=3,
})
before = time.now()
ok, msg = pcall(t.waitDialog, t, 100*time.millisecond)
after = time.now()
assert.eq(ok, false)
assert.eq(msg, "testdata/scenario/handle-dialog.lua:78: timeout")
assert.ge(after - before, 100*time.millisecond)
assert.lt(after - before, 200*time.millisecond)


called = false
t:onDialog(function(dialog)
    called = true
    assert.eq(dialog, {
        type    = "confirm",
        message = "are you sure?",
        url     = TEST.url("/dialog/confirm"),
    })
    return false
end)
assert.eq(called, false)
t:go(TEST.url("/dialog/confirm"))
assert.eq(called, true)
assert.eq(t("span").text, "false")


called = false
t:onDialog(function(dialog)
    called = true
    assert.eq(dialog, {
        type    = "prompt",
        message = "type something here!",
        url     = TEST.url("/dialog/prompt"),
    })
    return false
end)
assert.eq(called, false)
t:go(TEST.url("/dialog/prompt"))
assert.eq(called, true)
assert.eq(t("span").text, [[null]])


called = false
t:onDialog(function(dialog)
    called = true
    assert.eq(dialog, {
        type    = "prompt",
        message = "type something here!",
        url     = TEST.url("/dialog/prompt"),
    })
    return true, "hello"
end)
assert.eq(called, false)
t:go(TEST.url("/dialog/prompt"))
assert.eq(called, true)
assert.eq(t("span").text, [["hello"]])


t:onDialog(function(dialog) error("it should be disabled") end)
t:onDialog(nil)
t:go(TEST.url("/dialog/alert"))
