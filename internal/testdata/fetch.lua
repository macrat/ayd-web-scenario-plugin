resp = fetch(TEST.url("/header"))
assert.eq(
    resp,
    {
        headers = {
            Date               = resp.headers.Date,
            ["Content-Type"]   = {"text/plain; charset=utf-8"},
            ["Content-Length"] = {"6"},
        },
        url    = TEST.url("/header"),
        status = 200,
        length = 6,
        read   = resp.read,
    }
)
assert.eq(resp:read(), [[GET ""]])


resp = fetch(TEST.url("/header"), {method="POST", headers={["X-Header-Test"]="hello world"}})
assert.eq(
    resp,
    {
        headers = {
            Date               = resp.headers.Date,
            ["Content-Type"]   = {"text/plain; charset=utf-8"},
            ["Content-Length"] = {"18"},
        },
        url    = TEST.url("/header"),
        status = 200,
        length = 18,
        read   = resp.read,
    }
)
assert.eq(
    resp:read(),
    [[POST "hello world"]]
)


resp = fetch(TEST.url("/echo"), {body="hello"})
assert.eq(resp.status, 200)
assert.eq(resp:read(), [[hello]])

resp = fetch(TEST.url("/echo"), {body=123})
assert.eq(resp.status, 200)
assert.eq(resp:read(), [[123]])

reader_idx = 0
function reader()
    reader_idx = reader_idx + 1
    return ({
        "hello",
        "world",
    })[reader_idx]
end
resp = fetch(TEST.url("/echo"), {body=reader})
assert.eq(resp.status, 200)
assert.eq(resp:read(), "hello\nworld\n")


resp = fetch(TEST.url("/error"), {})
assert.eq(
    resp,
    {
        headers = {
            Date               = resp.headers.Date,
            ["Content-Type"]   = {"text/plain; charset=utf-8"},
            ["Content-Length"] = {"16"},
        },
        url    = TEST.url("/error"),
        status = 500,
        length = 16,
        read   = resp.read,
    }
)
assert.eq(
    resp:read(),
    [[something wrong!]]
)


ok, err = pcall(fetch, TEST.url("/slow"), {timeout=10*time.millisecond})
assert.eq(ok, false)
assert.eq(err, "testdata/fetch.lua:82: timeout")

ok = pcall(fetch, TEST.url("/slow"), {timeout=500*time.millisecond})
assert.eq(ok, true)


resp, jar = fetch(TEST.url("/cookie/get"))
assert.eq(resp.status, 200)
assert.eq(resp:read(), "not set")

assert.eq(jar:all(), {})

resp = fetch(TEST.url("/cookie/set"), {cookiejar=jar})
assert.eq(resp.status, 200)
assert.eq(resp:read(), "ok")

resp = fetch(TEST.url("/cookie/get"), {cookiejar=jar})
assert.eq(resp.status, 200)
assert.eq(resp:read(), "hello world")

resp = fetch(TEST.url("/cookie/get"))
assert.eq(resp.status, 200)
assert.eq(resp:read(), "not set")

assert.eq(jar:all(), {
    [TEST.url("/cookie/set")] = {{
        name     = "cookie_test",
        value    = "hello world",
        path     = "",
        domain   = "",
        secure   = false,
        httponly = false,
        samesite = "",
    }},
})

assert.eq(jar:get(TEST.url("/cookie/set")), {{
    name     = "cookie_test",
    value    = "hello world",
    path     = "",
    domain   = "",
    secure   = false,
    httponly = false,
    samesite = "",
}})
