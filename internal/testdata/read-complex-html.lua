t = tab.new(TEST.url("/complex-dom"))

assert.eq(t("div").text, "text\nhello beautiful world!")

assert.eq(t("b").text, "hello ")


texts = {}
for elm in t:all("b") do
    table.insert(texts, elm.text)
end
assert.eq(texts, {"hello ", "world"})


for elm in t:all("input[type=text]") do
    elm:sendKeys("def")
end
for elm in t:all("input[type=text]") do
    assert.eq(elm.value, "def")
end

t("input[type=text]"):setValue("abc")
assert.eq(#t:all("input[type=text]"), 2)
assert.eq(t:all("input[type=text]")[1].value, "abc")
assert.eq(t:all("input[type=text]")[2].value, "def")


function getall(itr, param)
    result = {}
    for elm in itr do
        table.insert(result, elm[param])
    end
    return table.concat(result, ":")
end

assert.eq(getall(t:all("h1"), "text"), "text:form")
assert.eq(t("div")("h1").text , "text")
assert.eq(t("form")("h1").text, "form")
assert.eq(getall(t("div"):all("h1"), "text"), "text")
assert.eq(getall(t("form"):all("h1"), "text"), "form")


assert.eq(t("form").action, "GET")
assert.eq(t("input").type, "text")
