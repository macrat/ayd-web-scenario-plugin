t = tab.new(TEST.url("/complex-dom"))

assert(t("div").text == "hello beautiful world!")

assert(t("b").text == "hello ")


assert(#t:all("b") == 2)
result = ""
for _, e in ipairs(t:all("b")) do
    result = result .. e.text
end
assert(result == "hello world")


texts = t:all("b").text
assert(#texts == 2)
assert(texts[1] == "hello ")
assert(texts[2] == "world")


t:all("input[type=text]"):sendKeys("def")
assert(table.concat(t:all("input[type=text]").value, ",") == "def,def")

t("input[type=text]"):setValue("abc")
assert(table.concat(t:all("input[type=text]").value, ",") == "abc,def")
