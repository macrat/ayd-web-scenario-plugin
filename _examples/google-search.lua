t = tab.new()
       :recording(true)
       :go("https://www.google.com")

t("input[type=text]")
    :sendKeys("ayd status monitoring tool")
    :submit()

time.sleep(1*time.second)

for i, elm in ipairs(t:all("h1")) do
    print.extra(string.format("result_%d", i), elm.text)
end

t:close()
