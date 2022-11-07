t = tab.new()


called = false
t:onResponse(function(resp)
    called = true

    assert.ne(resp.id, "")
    assert.lt(0, resp.remotePort)
    assert.le(resp.remotePort, 65535)
    assert.ne(resp.remoteIP, "")

    assert.eq(resp, {
        id         = resp.id,
        type       = "Document",
        url        = TEST.url("/"),
        status     = 200,
        length     = 116,
        remoteIP   = resp.remoteIP,
        remotePort = resp.remotePort,
        headers    = {
            Date               = resp.headers.Date,
            ["Content-Type"]   = "text/html; charset=utf-8",
            ["Content-Length"] = "86",
        },
    })

    assert.eq(resp:read("all"), '<title>world - test</title><div id="greeting">hello <b class="target">world</b>!</div>')
end)
t:go(TEST.url("/"))
assert.eq(called, true)


assert.eq(t.responses, {
    {
        id         = t.responses[1].id,
        type       = "Document",
        url        = TEST.url("/"),
        status     = 200,
        length     = 116,
        remoteIP   = "127.0.0.1",
        remotePort = t.responses[1].remotePort,
        headers    = {
            Date               = t.responses[1].headers.Date,
            ["Content-Type"]   = "text/html; charset=utf-8",
            ["Content-Length"] = "86",
        },
    },
    _waited=0,
})

for _, r in ipairs(t.responses) do
    _, w = t:waitResponse(0)
    assert.eq(r, w)
end
