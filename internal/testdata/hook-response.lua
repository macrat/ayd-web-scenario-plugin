t = tab.new()


called = false
t:onResponse(function(resp)
    called = true

    assert(resp.id ~= "")
    assert(resp.type == "Document")
    assert(resp.url == TEST.url("/"))
    assert(resp.status == 200)
    assert(resp.mimetype == "text/html")
    assert(resp.remoteIP ~= "")
    assert(0 < resp.remotePort and resp.remotePort <= 65535)
    assert(resp.length == 116)

    assert(resp:body() == '<title>world - test</title><div id="greeting">hello <b class="target">world</b>!</div>')
end)
t:go(TEST.url("/"))
assert(called == true)
