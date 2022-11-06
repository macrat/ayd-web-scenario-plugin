assert.eq(toxml({
    "hello", foo="bar", fizz="buzz",
    {"hoge", "fuga"},
    "wah",
    {"foo", "bar", 123, abc=456},
}), [[<hello foo="bar" fizz="buzz"><hoge>fuga</hoge>wah<foo abc="456">bar123</foo></hello>]])


assert.eq(fromxml([=[
    <hello foo="bar" fizz="buzz">
        <hoge>fuga</hoge>
        wah
        <foo abc="123"><![CDATA[bar
baz]]></foo>
    </hello>
]=]), {
    "hello", foo="bar", fizz="buzz",
    {"hoge", "fuga"},
    "wah",
    {"foo", "bar\nbaz", abc="123"},
})

assert.eq(fromxml({
    [=[ <hello foo="bar" fizz="buzz"> ]=],
    [=[ <hoge>fuga</hoge>             ]=],
    [=[ wah                           ]=],
    [=[ <foo abc="123"><![CDATA[bar   ]=],
    [=[ baz]]></foo>                  ]=],
    [=[ </hello>                      ]=],
}), {
    "hello", foo="bar", fizz="buzz",
    {"hoge", "fuga"},
    "wah",
    {"foo", "bar   \n baz", abc="123"},
})
