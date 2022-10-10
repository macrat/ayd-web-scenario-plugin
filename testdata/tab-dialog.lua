t = tab.new()


called = false
t:onDialog(function(typ, msg, url)
    called = true
    assert(typ == "alert")
    assert(msg == "welcome!")
    assert(url == TEST.url("/dialog/alert"))
    return true
end)
assert(called == false)
t:go(TEST.url("/dialog/alert"))
assert(called == true)


called = false
t:onDialog(function(typ, msg, url)
    called = true
    assert(typ == "confirm")
    assert(msg == "are you sure?")
    assert(url == TEST.url("/dialog/confirm"))
    return true
end)
assert(called == false)
t:go(TEST.url("/dialog/confirm"))
assert(called == true)
assert(t("span").text == "true")


called = false
t:onDialog(function(typ, msg, url)
    called = true
    assert(typ == "confirm")
    assert(msg == "are you sure?")
    assert(url == TEST.url("/dialog/confirm"))
    return false
end)
assert(called == false)
t:go(TEST.url("/dialog/confirm"))
assert(called == true)
assert(t("span").text == "false")


called = false
t:onDialog(function(typ, msg, url)
    called = true
    assert(typ == "prompt")
    assert(msg == "type something here!")
    assert(url == TEST.url("/dialog/prompt"))
    return false
end)
assert(called == false)
t:go(TEST.url("/dialog/prompt"))
assert(called == true)
assert(t("span").text == [[null]])


called = false
t:onDialog(function(typ, msg, url)
    called = true
    assert(typ == "prompt")
    assert(msg == "type something here!")
    assert(url == TEST.url("/dialog/prompt"))
    return true, "hello"
end)
assert(called == false)
t:go(TEST.url("/dialog/prompt"))
assert(called == true)
assert(t("span").text == [["hello"]])
