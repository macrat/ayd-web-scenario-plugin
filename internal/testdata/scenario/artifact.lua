assert.eq(artifact.path, TEST.storage())


assert.eq(artifact.list, {})

f, err = artifact.open("test.txt", "w")
assert.eq(err, nil)
f:write("hello world")
f:close()

assert.eq(artifact.list, {"test.txt"})

f = artifact.open("test.txt")
assert.eq(f:read("*a"), "hello world")
f:close()

assert.eq(artifact.list, {"test.txt"})

artifact.remove("test.txt")
assert.eq(artifact.list, {})

ok, err = pcall(artifact.remove, "test.txt")
assert.eq(ok, false)
assert.eq(err, "testdata/scenario/artifact.lua:22: file does not exist")
