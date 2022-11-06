t = tab.new(TEST.url())

assert.eq(t(".target").text, "world")
assert.eq(t("#greeting").text, "hello world!")

assert.eq(t(".target").innerHTML, [[world]])
assert.eq(t("#greeting").innerHTML, [[hello <b class="target">world</b>!]])

assert.eq(t(".target").outerHTML, [[<b class="target">world</b>]])
assert.eq(t("#greeting").outerHTML, [[<div id="greeting">hello <b class="target">world</b>!</div>]])

assert.eq(t(".target").class, "target")
assert.eq(t("#greeting").id, "greeting")

ok, err = pcall(t, "#no-such-element")
assert.eq(ok, false)
assert.eq(err, "testdata/scenario/read-simple-html.lua:15: no such element")

ok = pcall(t.wait, t, "#greeting", 100*time.millisecond)
assert.eq(ok, true)

ok, err = pcall(t.wait, t, "#no-such-element", 100*time.millisecond)
assert.eq(ok, false)
assert.eq(err, "testdata/scenario/read-simple-html.lua:22: timeout")
