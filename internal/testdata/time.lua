a = time.now()
b = time.now()

assert.le(a, b)
assert.gt(a+100, b)

time.sleep(200 * time.millisecond)
c = time.now()

assert.le(b+100, c)
assert.gt(b+300, c)

assert.eq(os.getenv("TZ"), "UTC")
assert.eq(time.format(1136214245999), "2006-01-02T15:04:05Z")
