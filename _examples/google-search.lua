t = tab.new()
       :recording(true)
       :go("https://www.google.com")

t("input[type=text]")
    :sendKeys("ayd status monitoring tool")
    :submit()

t:wait("div[role=main]")

for i, elm in ipairs(t:all("h1")) do
    print.extra(string.format("result_%d", i), elm.text)
end

t:xpath("//a[contains(h3[text()], 'macrat/ayd')]")[1]:click()

t:wait(".octicon")

t:close()
