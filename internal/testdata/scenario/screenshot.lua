package.path = package.path .. ";./testdata/?.lua"
storage = require("utils/storage")

t = tab.new(TEST.url())

assert.eq(storage.list(), {})

t:screenshot()
assert.eq(storage.list(), {"000001.png"})

t("b"):screenshot()
assert.eq(storage.list(), {"000001.png", "000002.png"})
