resp = fetch("https://status.aws.amazon.com/rss/ec2-us-east-1.rss")
assert.eq(resp.status, 200)

history = {}

rss = fromxml(resp:read("a"))
for _, item in ipairs(rss[2]) do
    if item[1] == "item" then
        info = {}
        for _, x in ipairs(item) do
            if type(x) == "table" then
                info[x[1]] = x[2]
            end
        end
        table.insert(history, info)
        print(info.title)

        if not string.match(info.title, "^Informational message:") and not string.match(info.title, "[RESOLVED]") then
            print.status("FAILURE")
        end
    end
end

print.extra("history", history)
