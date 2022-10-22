t = tab.new()

t2 = tab.new(TEST.url())
t2:onDialog(function(ev)
    error(string.format("caught unexpected dialog: %s %s %s", ev.type, ev.message, ev.url))
end)


called = false
t:onDialog(function(ev)
    called = true
    assert(ev.type == "alert")
    assert(ev.message == "welcome!")
    assert(ev.url == TEST.url("/dialog/alert"))
    return true
end)
assert(called == false)
t:go(TEST.url("/dialog/alert"))
assert(called == true)


called = false
t:onDialog(function(ev)
    called = true
    assert(ev.type == "confirm")
    assert(ev.message == "are you sure?")
    assert(ev.url == TEST.url("/dialog/confirm"))
    return true
end)
assert(called == false)
t:go(TEST.url("/dialog/confirm"))
assert(called == true)
assert(t("span").text == "true")


called = false
t:onDialog(function(ev)
    called = true
    assert(ev.type == "confirm")
    assert(ev.message == "are you sure?")
    assert(ev.url == TEST.url("/dialog/confirm"))
    return false
end)
assert(called == false)
t:go(TEST.url("/dialog/confirm"))
assert(called == true)
assert(t("span").text == "false")


called = false
t:onDialog(function(ev)
    called = true
    assert(ev.type == "prompt")
    assert(ev.message == "type something here!")
    assert(ev.url == TEST.url("/dialog/prompt"))
    return false
end)
assert(called == false)
t:go(TEST.url("/dialog/prompt"))
assert(called == true)
assert(t("span").text == [[null]])


called = false
t:onDialog(function(ev)
    called = true
    assert(ev.type == "prompt")
    assert(ev.message == "type something here!")
    assert(ev.url == TEST.url("/dialog/prompt"))
    return true, "hello"
end)
assert(called == false)
t:go(TEST.url("/dialog/prompt"))
assert(called == true)
assert(t("span").text == [["hello"]])
