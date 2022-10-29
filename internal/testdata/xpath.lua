t = tab.new(TEST.url("/complex-dom"))

xs = t:xpath("//b")
assert.eq(#xs, 2)
assert.eq(xs[1].text, "hello ")
assert.eq(xs[2].text, "world")

xs = t:xpath("//div/b[1]")
assert.eq(#xs, 1)
assert.eq(xs[1].text, "hello ")

xs = t:xpath("//div/b[text()='world']")
assert.eq(#xs, 1)
assert.eq(xs[1].class, "target")
