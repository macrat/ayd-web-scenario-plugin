t = tab.new("https://google.com")

t("input[type=text]")
    :sendKeys("ayd status monitoring")
    :submit()

time.sleep(500)

t:screenshot("screenshot.jpg")

print(t.title)
