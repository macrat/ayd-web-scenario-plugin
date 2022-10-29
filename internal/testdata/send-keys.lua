t = tab.new(TEST.url("/dynamic"))
e = t("#key-listener")

assert(e:sendKeys("x").text == 'x')
assert(e:sendKeys("X").text == 'shift+X')
assert(e:sendKeys("x", {"shift"}).text == 'shift+x')
assert(e:sendKeys("a", {"alt"}).text == 'alt+a')
assert(e:sendKeys("c", {"ctrl"}).text == 'ctrl+c')
assert(e:sendKeys("m", {"meta"}).text == 'meta+m')
assert(e:sendKeys("s", {"shift"}).text == 'shift+s')
assert(e:sendKeys("x", {"ctrl", "shift"}).text == 'ctrl+shift+x')
assert(e:sendKeys("X", {"alt", "meta"}).text == 'alt+meta+shift+X')

assert(e:sendKeys(key.f2).text == 'F2')
assert(e:sendKeys(key.backspace).text == 'Backspace')

assert(
    t("input[type=text]")
        :sendKeys("helle" .. key.backspace .. "o world")
        .value == "hello world"
)
