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
        mimetype   = "text/html",
        remoteIP   = resp.remoteIP,
        remotePort = resp.remotePort,
        length     = 116,
        body       = resp.body,
    })

    assert.eq(resp:body(), '<title>world - test</title><div id="greeting">hello <b class="target">world</b>!</div>')
end)
t:go(TEST.url("/"))
assert.eq(called, true)


assert.eq(t.responses, {
    {
        id         = t.responses[1].id,
        type       = "Document",
        url        = TEST.url("/"),
        status     = 200,
        mimetype   = "text/html",
        remoteIP   = "127.0.0.1",
        remotePort = t.responses[1].remotePort,
        length     = 116,
        body       = t.responses[1].body,
    },
    _waited=0,
})

for _, r in ipairs(t.responses) do
    _, w = t:waitResponse(0)
    assert.eq(r, w)
end
