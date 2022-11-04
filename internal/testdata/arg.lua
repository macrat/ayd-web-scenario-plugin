assert.eq(arg, {
    "abc",
    "def",
    target = {
        url      = "web-scenario://foo:xxxxx@/dummy/script.lua?hello=world&hoge=fuga#piyo",
        username = "foo",
        password = arg.target.password,
        query    = {hello="world", hoge="fuga"},
        fragment = "piyo",
    },
    debug     = false,
    head      = false,
    recording = false,
})
assert.eq(arg.target.password(), "bar")
