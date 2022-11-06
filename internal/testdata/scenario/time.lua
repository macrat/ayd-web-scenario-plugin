a = time.now()
b = time.now()

assert.le(a, b)
assert.lt(b-a, 200*time.millisecond)

c = time.now()
time.sleep(100 * time.millisecond)
d = time.now()

assert.ge(d-c, 100*time.millisecond)
assert.lt(d-c, 300*time.millisecond)

assert.eq(os.getenv("TZ"), "UTC")
assert.eq(time.format(1136214245000), "2006-01-02T15:04:05+0000")
assert.eq(time.format(1136214245000, "%Y/%m/%d"), "2006/01/02")
assert.eq(time.format(1136214245000, "%H:%M:%S"), "15:04:05")
