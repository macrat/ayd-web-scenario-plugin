t = tab.new("https://www.google.com")

t("input[type=text]")
    :sendKeys("ayd status monitoring tool")
    :submit()

time.sleep(1*time.second)

for i, x in ipairs(t:all("h1").text) do
    print.extra(i, x)
end
