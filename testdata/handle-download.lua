t = tab.new(TEST.url("/download"))


t("a"):click()

f = t:waitDownload()
assert(TEST.storage("data.txt") == f.path)
assert(f.bytes == 14)

assert(io.popen("dir " .. TEST.storage()):read() == "data.txt")
assert(io.input(TEST.storage("data.txt")):read() == "this is a data")


t("a"):click()

f = t:waitDownload()
assert(TEST.storage("data.txt") == f.path)
assert(f.bytes == 14)

assert(io.popen("dir " .. TEST.storage()):read() == "data.txt")
assert(io.input(TEST.storage("data.txt")):read() == "this is a data")
