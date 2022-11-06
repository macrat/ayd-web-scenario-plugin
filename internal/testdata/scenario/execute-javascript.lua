t = tab.new(TEST.url())

assert.eq(t:eval([[ "hello world" ]]), "hello world")
assert.eq(t:eval([[ 1 + 2 ]]), 3)

assert.eq(
    t:eval([[
        document.querySelector("#greeting").innerText.replace(/[^a-z]/g, "")
    ]]),
    "helloworld"
)

t:eval([[ document.querySelector(".target").innerText = "eval" ]])
assert.eq(t("#greeting").text, "hello eval!")
