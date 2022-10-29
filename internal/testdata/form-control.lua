t = tab.new(TEST.url("/dynamic"))


assert.eq(t("ol").innerHTML, "")
t("#append"):click()
assert.eq(t("ol li:last-child").outerHTML, "<li>count=0</li>")
t("#append"):click()
assert.eq(t("ol li:last-child").outerHTML, "<li>count=1</li>")


assert.eq(t("#text").text, "")

t("input[type=text]")
    :sendKeys("hello")
    :sendKeys(" ")
    :sendKeys("world")
    :blur()
assert.eq(t("#text").text, "hello world")
assert.eq(t("input[type=text]").value, "hello world")

t("input[type=text]")
    :setValue("")
    :sendKeys("あいうえお")
    :blur()
assert.eq(t("#text").text, "あいうえお")
assert.eq(t("input[type=text]").value, "あいうえお")


assert.eq(t("#look-at-me").text, "blur")
assert.eq(t("#look-at-me"):focus().text, "focus")
assert.eq(t("#look-at-me"):blur().text, "blur")


assert.eq(t("#submitted").text, "")
t("textarea")
    :sendKeys("hello webscenario")
    :submit()
assert.eq(t("#submitted").text, "hello webscenario")
assert.eq(t.url, TEST.url("/dynamic?textarea=hello+webscenario"))
