t = tab.new(TEST.url())
assert.eq(t("body").text, "hello world!")

t:go(TEST.url("/?target=foo"))
assert.eq(t("body").text, "hello foo!")

t:go(TEST.url("/?target=bar"))
assert.eq(t("body").text, "hello bar!")

t:back()
assert.eq(t("body").text, "hello foo!")

t:back()
assert.eq(t("body").text, "hello world!")

t:forward()
assert.eq(t("body").text, "hello foo!")

t:go(TEST.url("/counter"))
before = tonumber(t("span").text)
assert.gt(before, 0)
t:reload()
after = tonumber(t("span").text)
assert.gt(after, 0)
assert.gt(after, before)
