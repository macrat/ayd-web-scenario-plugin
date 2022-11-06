t = tab.new(TEST.url("/dynamic"))


e = t("#click-listener")
assert.eq(e:click().text, "0")
assert.eq(e:click("right").text, "2")
assert.eq(e:click("left").text, "0")
assert.eq(e:click("middle").text, "1")


t:go(TEST.url("/"))
t:wait("#greeting")
assert.eq(t.url, TEST.url("/"))

t("body"):click("back")
t:wait("#click-listener")
assert.eq(t.url, TEST.url("/dynamic"))

t("body"):click("forward")
t:wait("#greeting")
assert.eq(t.url, TEST.url("/"))
