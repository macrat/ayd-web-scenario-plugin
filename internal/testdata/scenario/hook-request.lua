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


assert.eq(t.requests, {
    {id=t.requests[1].id, method="GET",  type="Document", url=TEST.url("/"),   },
    {id=t.requests[2].id, method="POST", type="Document", url=TEST.url("/post"), body="value=hello+POST+form"},
    _waited=0,
})

for _, r in ipairs(t.requests) do
    _, w = t:waitRequest(0)
    assert.eq(r, w)
end
