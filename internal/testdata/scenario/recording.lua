t = tab.new({
    url=TEST.url("/"),
    recording=true,
    width=300,
    height=300
})

t:go(TEST.url("/dynamic"))
t("button"):click():click()

t("textarea"):sendKeys("hello world")
t("input[type=submit]"):click()

t:close()

while #artifact.list < 1 do
    time.sleep(100*time.millisecond)
end
assert.eq(artifact.list, {"record1.gif"})

-- loop to wait for writing GIF.
while #artifact.open("record1.gif", "rb"):read("*a") == 0 do
    time.sleep(100*time.millisecond)
end
time.sleep(100*time.millisecond)

-- generate test data
--print(io.popen("cp " .. artifact.path .. "/record1.gif testdata/gif/record-actual.gif"):read("*a"))

want = io.open("testdata/gif/record.gif", "rb"):read("*a")
assert(artifact.open("record1.gif", "rb"):read("*a") == want, "recorded gif is different")
