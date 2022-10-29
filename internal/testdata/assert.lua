function shouldError(fn, a, b, pattern)
    ok, err = pcall(function() fn(a, b) end)

    if ok then
        error("expected error but not happened", 3)
    end
    if not string.match(err, pattern) then
        error(string.format("expected error pattern was %q but got %q", pattern, err), 3)
    end
end

assert(true)
a, b = assert(true, "message")
assert.eq({a, b}, {true, "message"})
shouldError(assert, false, nil, ": assertion failed!")
shouldError(assert, false, "hello world", ": hello world")

a, b = assert.eq("hello", "hello")
assert.eq({a, b}, {"hello", "hello"})
assert.eq(123, 123.0)
assert.eq({"hello", 123}, {"hello", 123})
assert.eq({a=1, b=2}, {b=2, a=1})
shouldError(assert.eq, 123, 234, ": assertion failed: 123 == 234$")
shouldError(assert.eq, "hello", "world", ': assertion failed: "hello" == "world"$')
shouldError(assert.eq, {"hello", "world"}, "wah", ': assertion failed: {"hello", "world"} == "wah"$')

a, b = assert.ne("hello", "world")
assert.eq({a, b}, {"hello", "world"})
assert.ne(123, 124)
assert.ne({"hello", {a=123}}, {"hello", {a=124}})
shouldError(assert.ne, 123, 123, ": assertion failed: 123 ~= 123$")
shouldError(assert.ne, "hello", "hello", ': assertion failed: "hello" ~= "hello"$')

assert.lt(123, 124)
a, b = assert.lt("abc", "bcd")
assert.eq({a, b}, {"abc", "bcd"})
shouldError(assert.lt, 123, 123, ": assertion failed: 123 < 123$")
shouldError(assert.lt, 124, 123, ": assertion failed: 124 < 123$")
shouldError(assert.lt, "bcd", "abc", ': assertion failed: "bcd" < "abc"$')
shouldError(assert.lt, "abc", "abc", ': assertion failed: "abc" < "abc"$')

assert.le(123, 124)
assert.le(123, 123)
assert.le("abc", "abc")
assert.le("abc", "bcd")
shouldError(assert.le, 124, 123, ": assertion failed: 124 <= 123$")
shouldError(assert.le, "abd", "abc", ': assertion failed: "abd" <= "abc"$')

assert.gt(124, 123)
assert.gt("def", "abc")
shouldError(assert.gt, 123, 123, ": assertion failed: 123 > 123$")
shouldError(assert.gt, 123, 124, ": assertion failed: 123 > 124$")
shouldError(assert.gt, "abc", "def", ': assertion failed: "abc" > "def"$')
shouldError(assert.gt, "abc", "abc", ': assertion failed: "abc" > "abc"$')

assert.ge(124, 123)
assert.ge(123, 123)
assert.ge("def", "abc")
assert.ge("abc", "abc")
shouldError(assert.ge, 123, 124, ": assertion failed: 123 >= 124$")
shouldError(assert.ge, "abc", "dbc", ': assertion failed: "abc" >= "dbc"$')
