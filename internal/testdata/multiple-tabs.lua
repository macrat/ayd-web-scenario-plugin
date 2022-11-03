t1 = tab.new(TEST.url("/download"))
t2 = tab.new(TEST.url("/download"))

assert.eq(#t1.downloads, 0)
assert.eq(#t2.downloads, 0)

t1("a"):click()
t2("a"):click()

t1:waitDownload()
t2:waitDownload()

assert.eq(#t1.downloads, 1)
assert.eq(#t2.downloads, 1)
