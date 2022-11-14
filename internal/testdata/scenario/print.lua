ok = pcall(print.extra, "foo", "bar")
assert(ok, "failed to set foo")

for _, key in ipairs({"time", "target"}) do
    ok, err = pcall(print.extra, key, 123)
    assert(not ok, "unexpectedly succeed to set " .. key)
    assert.eq(err, "testdata/scenario/print.lua:5: print.extra() can not set " .. key .. ".")
end

for key, alt in ipairs({message="print()", latency="print.latency()", status="print.status()"}) do
    ok, err = pcall(print.extra, key, 123)
    assert(not ok, "unexpectedly succeed to set " .. key)
    assert.eq(err, "testdata/scenario/print.lua:11: print.extra() can not set " .. key .. ". please use " .. alt .. ".")
end
