function collect(iter)
    xs = {}
    for x in iter do
        table.insert(xs, x)
    end
    return xs
end


csv, header = fromcsv([[hello,foo,hello
world,bar,same
,baz,name
csv,,
]])
assert.eq(header, {"hello", "foo", "hello_1"})
assert.eq(collect(csv), {
    {hello="world", foo="bar", hello_1="same"},
    {hello="", foo="baz", hello_1="name"},
    {hello="csv", foo="", hello_1=""},
})

csv, header = fromcsv({
    "hello,foo",
    "world,bar",
    ",baz",
    "csv,"
})
assert.eq(header, {"hello", "foo"})
assert.eq(collect(csv), {
    {hello="world", foo="bar"},
    {hello="", foo="baz"},
    {hello="csv", foo=""},
})

i = 0
function f()
    i = i + 1
    return ({
        "hello,foo",
        "world,bar",
        ",baz",
        "csv,"
    })[i]
end
csv, header = fromcsv(f)
assert.eq(header, {"hello", "foo"})
assert.eq(collect(csv), {
    {hello="world", foo="bar"},
    {hello="", foo="baz"},
    {hello="csv", foo=""},
})


assert.eq(collect(tocsv({
    {hello="world", foo="bar"},
    {foo="baz"},
    {hello="csv", nah="wah"},
}, {"hello", "foo"})), {"hello,foo", "world,bar", ",baz", "csv,"})

assert.eq(collect(tocsv({
    {hello="world"},
    {foo="baz"},
    {hello="csv", nah="wah"},
}, {"hello", "foo"})), {"hello,foo", "world,", ",baz", "csv,"})

assert.eq(collect(tocsv({
    {"world", "bar"},
    {nil, "baz"},
    {"csv", "wah"},
}, {"hello", "foo"})), {"hello,foo", "world,bar", ",baz", "csv,wah"})

assert.eq(collect(tocsv({
    {"world", "bar"},
    {"", "baz"},
    {"csv", "wah"},
}, false)), {"world,bar", ",baz", "csv,wah"})

assert.eq(collect(tocsv({
    {hello="world", foo="bar"},
    {foo="baz"},
    {hello="csv", nah="wah"},
})), {"foo,hello", "bar,world", "baz,", ",csv"})

assert.eq(collect(tocsv({
    {hello="world"},
    {foo="baz"},
    {hello="csv", nah="wah"},
}), true), {"hello", "world", "", "csv"})

assert.eq(collect(tocsv({
    {hello="world"},
    {foo="baz"},
    {hello="csv", nah="wah"},
})), {"hello", "world", "", "csv"})

i = 0
function f()
    i = i + 1
    return ({
        {hello="world", foo="bar"},
        {foo="baz"},
        {hello="csv", nah="wah"},
    })[i]
end
assert.eq(
    collect(tocsv(f, {"hello", "foo"})),
    {"hello,foo", "world,bar", ",baz", "csv,"}
)


assert.eq(
    collect(tocsv(fromcsv(io.open("./testdata/encoding/iris.csv"):lines()))),
    collect(io.open("./testdata/encoding/iris.csv"):lines())
)
