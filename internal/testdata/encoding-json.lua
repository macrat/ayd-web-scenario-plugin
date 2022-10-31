assert.eq(fromjson([[
	1
]]), 1)

assert.eq(fromjson([[
	"hello"
]]), "hello")

assert.eq(fromjson([[
	[1, "hello", 3]
]]), {1, "hello", 3})

assert.eq(fromjson([[
	{"hello": ["json", "world"], "foo": "bar"}
]]), {hello={"json", "world"}, foo="bar"})


assert.eq(tojson(1), "1")

assert.eq(tojson("hello"), '"hello"')

assert.eq(tojson({1, "hello", 3}), '[1,"hello",3]')

assert.eq(tojson({hello={"json", "world"}}), '{"hello":["json","world"]}')
