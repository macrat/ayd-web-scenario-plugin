package.path = package.path .. ";./testdata/?.lua"
storage = require("utils/storage")

t1 = tab.new(TEST.url("/download"))
t2 = tab.new(TEST.url("/download"))

t1:onDownload(function(file)
    assert.eq(file, {
        path  = TEST.storage("data.txt"),
        bytes = 14,
    })
end)

t2:onDownload(function(file)
    error("should not reach here")
end)

t1("a")
    :click()
    :click()

_, file = t1:waitDownload()
assert.eq(file, {path=TEST.storage("data.txt"), bytes=14})
t1:waitDownload()

ok, err = pcall(t1.waitDownload, t1, 0)
assert.eq(ok, false)
assert.eq(err, "testdata/scenario/handle-download.lua:26: timeout")

assert.eq(t1.downloads, {
    {path=TEST.storage("data.txt"), bytes=14},
    {path=TEST.storage("data.txt"), bytes=14},
    _waited=2,
})
assert.eq(string.gsub(io.popen("ls " .. TEST.storage()):read("*a"), "\r\n", "\n"), "data.txt\n")

f = io.input(TEST.storage("data.txt"))
assert.eq(f:read("*a"), "this is a data")
f:close()
