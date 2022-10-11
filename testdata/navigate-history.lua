t = tab.new(TEST.url())
assert(t("body").text == "hello world!")

t:go(TEST.url("/?target=foo"))
assert(t("body").text == "hello foo!")

t:go(TEST.url("/?target=bar"))
assert(t("body").text == "hello bar!")

t:back()
assert(t("body").text == "hello foo!")

t:back()
assert(t("body").text == "hello world!")

t:forward()
assert(t("body").text == "hello foo!")

t:go(TEST.url("/counter"))
before = tonumber(t("span").text)
assert(before > 0)
t:reload()
after = tonumber(t("span").text)
assert(after > 0)
assert(after > before)
