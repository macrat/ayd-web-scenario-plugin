t = tab.new(TEST.url("/dynamic"))
e = t("#key-listener")

assert.eq(e:sendKeys("x").text, 'x')
assert.eq(e:sendKeys("X").text, 'shift+X')
assert.eq(e:sendKeys("x", {"shift"}).text, 'shift+x')
assert.eq(e:sendKeys("a", {"alt"}).text, 'alt+a')
assert.eq(e:sendKeys("c", {"ctrl"}).text, 'ctrl+c')
assert.eq(e:sendKeys("m", {"meta"}).text, 'meta+m')
assert.eq(e:sendKeys("s", {"shift"}).text, 'shift+s')
assert.eq(e:sendKeys("x", {"ctrl", "shift"}).text, 'ctrl+shift+x')
assert.eq(e:sendKeys("X", {"alt", "meta"}).text, 'alt+meta+shift+X')

assert.eq(e:sendKeys(key.f2).text, 'F2')
assert.eq(e:sendKeys(key.backspace).text, 'Backspace')

assert.eq(
    t("input[type=text]")
        :sendKeys("helle" .. key.backspace .. "o world")
        .value,
    "hello world"
)
