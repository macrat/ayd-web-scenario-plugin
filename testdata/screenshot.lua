t = tab.new(TEST.url())

assert(io.popen("ls " .. TEST.storage()):read("*a") == "")

t:screenshot()
assert(io.popen("ls " .. TEST.storage()):read("*a") == "000001.jpg\n")

t("b"):screenshot()
assert(io.popen("ls " .. TEST.storage()):read("*a") == "000001.jpg\n000002.jpg\n")

-- XXX: This test works on only UNIX. It doesn't work on Windows.
