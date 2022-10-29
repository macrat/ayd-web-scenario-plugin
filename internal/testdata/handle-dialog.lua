t = tab.new()
t:go(TEST.url("/dialog/alert"))

t2 = tab.new(TEST.url())
t2:onDialog(function(dialog)
    error(string.format("caught unexpected dialog: %s %s %s", dialog.type, dialog.message, dialog.url))
end)


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
