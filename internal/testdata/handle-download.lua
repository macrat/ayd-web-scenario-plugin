package.path = package.path .. ";./testdata/?.lua"
storage = require("utils/storage")

t1 = tab.new(TEST.url("/download"))
t2 = tab.new(TEST.url("/download"))

download_count = 0
t1:onDownloaded(function(file)
    assert.eq(file, {
        path  = TEST.storage("data.txt"),
        bytes = 14,
    })
    download_count = download_count + 1
end)

t2:onDownloaded(function(file)
    error("should not reach here")
end)

t1("a")
    :click()
    :click()

while download_count < 2 do
    time.sleep(5 * time.millisecond)
end

assert.eq(download_count, 2)
assert.eq(string.gsub(io.popen("ls " .. TEST.storage()):read("*a"), "\r\n", "\n"), "data.txt\n")

f = io.input(TEST.storage("data.txt"))
assert.eq(f:read("*a"), "this is a data")
f:close()
