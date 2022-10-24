t = tab.new(TEST.url("/complex-dom"))

assert(t("div").text == "text\nhello beautiful world!")

assert(t("b").text == "hello ")


texts = {}
for elm in t:all("b") do
    table.insert(texts, elm.text)
end
print(#texts)
assert(#texts == 2)
assert(texts[1] == "hello ")
assert(texts[2] == "world")


for elm in t:all("input[type=text]") do
    elm:sendKeys("def")
end
for elm in t:all("input[type=text]") do
    assert(elm.value == "def")
end

t("input[type=text]"):setValue("abc")
assert(#t:all("input[type=text]") == 2)
assert(t:all("input[type=text]")[1].value == "abc")
assert(t:all("input[type=text]")[2].value == "def")


function getall(itr, param)
    result = {}
    for elm in itr do
        table.insert(result, elm[param])
    end
    return table.concat(result, ":")
end

assert(getall(t:all("h1"), "text") == "text:form")
assert(t("div")("h1").text == "text")
assert(t("form")("h1").text == "form")
assert(getall(t("div"):all("h1"), "text") == "text")
assert(getall(t("form"):all("h1"), "text") == "form")


assert(t("form").action == "GET")
assert(t("input").type == "text")
