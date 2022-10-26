t = tab.new()


called = false
t:onRequest(function(req)
    called = true
    assert(req.id ~= "")
    assert(req.type == "Document")
    assert(req.url == TEST.url("/"))
    assert(req.method == "GET")
    assert(req.body == nil)
end)
t:go(TEST.url("/"))
assert(called == true)


t:onRequest(nil)
t:go(TEST.url("/post"))
t("input[name=value]"):sendKeys("hello POST form")

called = false
t:onRequest(function(req)
    called = true

    assert(req.id ~= "")
    assert(req.type == "Document")
    assert(req.url == TEST.url("/post"))
    assert(req.method == "POST")
    assert(req.body == "value=hello+POST+form")
end)
t("input[type=submit]"):click()
assert(t("span").text == "hello POST form")
assert(called == true)
