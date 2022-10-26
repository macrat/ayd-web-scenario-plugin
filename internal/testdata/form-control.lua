t = tab.new(TEST.url("/dynamic"))


assert(t("ol").innerHTML == "")
t("#append"):click()
assert(t("ol li:last-child").outerHTML == "<li>count=0</li>")
t("#append"):click()
assert(t("ol li:last-child").outerHTML == "<li>count=1</li>")


assert(t("#text").text == "")

t("input[type=text]")
    :sendKeys("hello")
    :sendKeys(" ")
    :sendKeys("world")
    :blur()
assert(t("#text").text == "hello world")
assert(t("input[type=text]").value == "hello world")

t("input[type=text]")
    :setValue("")
    :sendKeys("あいうえお")
    :blur()
assert(t("#text").text == "あいうえお")
assert(t("input[type=text]").value == "あいうえお")


assert(t("#look-at-me").text == "blur")
assert(t("#look-at-me"):focus().text == "focus")
assert(t("#look-at-me"):blur().text == "blur")


assert(t("#submitted").text == "")
t("textarea")
    :sendKeys("hello webscenario")
    :submit()
assert(t("#submitted").text == "hello webscenario")
assert(t.url == TEST.url("/dynamic?textarea=hello+webscenario"))
