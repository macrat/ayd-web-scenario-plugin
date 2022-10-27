t = tab.new(TEST.url("/complex-dom"))

xs = t:xpath("//b")
assert(#xs == 2)
assert(xs[1].text == "hello ")
assert(xs[2].text == "world")

xs = t:xpath("//div/b[1]")
assert(#xs == 1)
assert(xs[1].text == "hello ")

xs = t:xpath("//div/b[text()='world']")
assert(#xs == 1)
assert(xs[1].class == "target")
