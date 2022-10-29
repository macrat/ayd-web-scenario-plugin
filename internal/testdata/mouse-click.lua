t = tab.new(TEST.url("/dynamic"))


e = t("#click-listener")
assert(e:click().text == "0", e.text)
assert(e:click("right").text == "2")
assert(e:click("left").text == "0")
assert(e:click("middle").text == "1")


t:go(TEST.url("/"))
t:wait("#greeting")
assert(t.url == TEST.url("/"))

t("body"):click("back")
t:wait("#click-listener")
assert(t.url == TEST.url("/dynamic"))

t("body"):click("forward")
t:wait("#greeting")
assert(t.url == TEST.url("/"))
