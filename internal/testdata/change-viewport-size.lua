t = tab.new(TEST.url("/window-size"))

assert(t.viewport.width == 1280)
assert(t.viewport.height == 720)
assert(t("body").text == "1280x720")

t:setViewport(1024, 768)
time.sleep(100)

assert(t.viewport.width == 1024)
assert(t.viewport.height == 768)
assert(t("body").text == "1024x768")
