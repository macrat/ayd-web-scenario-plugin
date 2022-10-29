t = tab.new(TEST.url("/window-size"))

assert.eq(t.viewport, {width=1280, height=720})
assert.eq(t("body").text, "1280x720")

t:setViewport(1024, 768)
time.sleep(100)

assert.eq(t.viewport, {width=1024, height=768})
assert.eq(t("body").text, "1024x768")
