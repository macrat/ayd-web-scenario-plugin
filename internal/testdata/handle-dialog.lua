t = tab.new()
t:go(TEST.url("/dialog/alert"))

t2 = tab.new(TEST.url())
t2:onDialog(function(dialog)
    error(string.format("caught unexpected dialog: %s %s %s", dialog.type, dialog.message, dialog.url))
end)


called = false
t:onDialog(function(dialog)
    called = true
    assert(dialog.type == "alert")
    assert(dialog.message == "welcome!")
    assert(dialog.url == TEST.url("/dialog/alert"))
    return true
end)
assert(called == false)
t:go(TEST.url("/dialog/alert"))
assert(called == true)


called = false
t:onDialog(function(dialog)
    called = true
    assert(dialog.type == "confirm")
    assert(dialog.message == "are you sure?")
    assert(dialog.url == TEST.url("/dialog/confirm"))
    return true
end)
assert(called == false)
t:go(TEST.url("/dialog/confirm"))
assert(called == true)
assert(t("span").text == "true")


called = false
t:onDialog(function(dialog)
    called = true
    assert(dialog.type == "confirm")
    assert(dialog.message == "are you sure?")
    assert(dialog.url == TEST.url("/dialog/confirm"))
    return false
end)
assert(called == false)
t:go(TEST.url("/dialog/confirm"))
assert(called == true)
assert(t("span").text == "false")


called = false
t:onDialog(function(dialog)
    called = true
    assert(dialog.type == "prompt")
    assert(dialog.message == "type something here!")
    assert(dialog.url == TEST.url("/dialog/prompt"))
    return false
end)
assert(called == false)
t:go(TEST.url("/dialog/prompt"))
assert(called == true)
assert(t("span").text == [[null]])


called = false
t:onDialog(function(dialog)
    called = true
    assert(dialog.type == "prompt")
    assert(dialog.message == "type something here!")
    assert(dialog.url == TEST.url("/dialog/prompt"))
    return true, "hello"
end)
assert(called == false)
t:go(TEST.url("/dialog/prompt"))
assert(called == true)
assert(t("span").text == [["hello"]])


t:onDialog(function(dialog) error("it should be disabled") end)
t:onDialog(nil)
t:go(TEST.url("/dialog/alert"))
