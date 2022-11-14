t = tab.new(TEST.url("/download"))

t:onDownload(function(ev)
    error("test error")
end)

t("a"):click()

t:waitDownload()

t:close()
