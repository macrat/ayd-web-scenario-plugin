t = tab.new(TEST.url())

assert(t:eval([[ "hello world" ]]) == "hello world")
assert(t:eval([[ 1 + 2 ]]) == 3)

assert(t:eval([[
    document.querySelector("#greeting").innerText.replace(/[^a-z]/g, "")
]]) == "helloworld")

t:eval([[ document.querySelector(".target").innerText = "eval" ]])
assert(t("#greeting").text == "hello eval!")
