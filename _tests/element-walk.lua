t = tab.new(TEST.url("/complex-dom"))

assert(t("div").text == "hello beautiful world!")

assert(t("b").text == "hello ")

assert(#t:all("b") == 2)
result = ""
for _, e in ipairs(t:all("b")) do
    result = result .. e.text
end
assert(result == "hello world")
