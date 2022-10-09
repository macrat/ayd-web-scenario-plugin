t = tab.new("https://githubstatus.com")
       :wait(".components-container")

t("body"):screenshot("githubstatus.com")

for _, elm in ipairs(t:all(".component-container:not([style*=\"display: none\"]) .component-inner-container")) do
    local name = elm(".name").text
    local status = elm(".status-msg").text
    print(name .. " is " .. status)

    assert(status == "Normal", name .. " is not operational!")
end
