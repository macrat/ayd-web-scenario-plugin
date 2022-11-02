local storage = {}

storage.list = function()
    xs = {}
    for x in io.popen("ls " .. TEST.storage()):lines() do
        table.insert(xs, x)
    end
    return xs
end

return storage
