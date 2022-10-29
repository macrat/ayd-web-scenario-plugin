t = tab.new(TEST.url())

assert.eq(io.popen("ls " .. TEST.storage()):read("*a"), "")

t:screenshot()
assert.eq(io.popen("ls " .. TEST.storage()):read("*a"), "000001.jpg\n")

t("b"):screenshot()
assert.eq(io.popen("ls " .. TEST.storage()):read("*a"), "000001.jpg\n000002.jpg\n")

-- XXX: This test works on only UNIX. It doesn't work on Windows.
