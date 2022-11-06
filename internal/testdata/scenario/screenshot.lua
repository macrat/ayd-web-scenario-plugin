t = tab.new(TEST.url())

assert.eq(artifact.list, {})

t:screenshot()
assert.eq(artifact.list, {"000001.png"})

t("b"):screenshot()
assert.eq(artifact.list, {"000001.png", "000002.png"})

for _, name in ipairs(artifact.list) do
    --io.open("testdata/screenshot/" .. name, "wb"):write(artifact.open(name, "rb"):read("*a"))

    assert.eq(
        io.open("testdata/screenshot/" .. name, "rb"):read("*a"),
        artifact.open(name, "rb"):read("*a")
    )
end
