a = time.now()
b = time.now()

assert(a <= b)
assert(a+100 > b)

time.sleep(200 * time.millisecond)
c = time.now()

assert(b+100 <= c)
assert(b+300 > c)

assert(os.getenv("TZ") == "UTC")
assert(time.format(1136214245999) == "2006-01-02T15:04:05Z", time.format(1136214245999))
