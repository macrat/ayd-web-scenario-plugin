t = tab.new()


called = false
t:onRequest(function(req)
    called = true
    assert.ne(req.id, "")

    assert.eq(req, {
        id     = req.id,
        type   = "Document",
        url    = TEST.url("/"),
        method = "GET",
        body   = nil,
    })
end)
t:go(TEST.url("/"))
assert.eq(called, true)


t:onRequest(nil)
t:go(TEST.url("/post"))
t("input[name=value]"):sendKeys("hello POST form")

called = false
t:onRequest(function(req)
    called = true

    assert.ne(req.id, "")

    assert.eq(req, {
        id     = req.id,
        type   = "Document",
        url    = TEST.url("/post"),
        method = "POST",
        body   = "value=hello+POST+form",
    })
end)
t("input[type=submit]"):click()
assert.eq(t("span").text, "hello POST form")
assert.eq(called, true)
