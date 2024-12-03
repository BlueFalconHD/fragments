function blogpost(title, date, author, description)
    return  "<a href='posts/" .. title .. "'><div class='blogpost'>\n" ..
            "   <h1>" .. title .. "</h1>\n" ..
            "   <p>" .. description .. "</p>\n" ..
            "   <p>" .. date .. " - " .. author .. "</p>\n" ..
            "</div></a>\n"
end

this:addBuilders {
    blogpostList = function()
        p = {}

        for file in io.popen('ls exampleSite/page/posts'):lines() do
            local f = fragments:get("../page/posts/" .. string.sub(file, 1, -6))
            table.insert(p, {
                title = f:getSharedMeta("postTitle"),
                date = f:getSharedMeta("postDate"),
                author = f:getSharedMeta("author"),
                description = f:getSharedMeta("postDescription")
            })
        end

        local result = ""
        for i, v in ipairs(p) do
            result = result .. blogpost(v.title, v.date, v.author, v.description)
        end

        return result
    end
}


---

*{blogpostList}
