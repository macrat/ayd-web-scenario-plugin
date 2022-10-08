t = tab.new(TEST.url())

assert(t(".target").text == "world")
assert(t("#greeting").text == "hello world!")

assert(t(".target").innerHTML == [[world]])
assert(t("#greeting").innerHTML == [[hello <b class="target">world</b>!]])

assert(t(".target").outerHTML == [[<b class="target">world</b>]])
assert(t("#greeting").outerHTML == [[<div id="greeting">hello <b class="target">world</b>!</div>]])
