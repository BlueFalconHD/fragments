function blogpost(id, title, date, author, description)
    return  "<a href='posts/" .. id .. ".html'><div class='blogpost'>\n" ..
            "   <h1>" .. title .. "</h1>\n" ..
            "   <p>" .. description .. "</p>\n" ..
            "   <p>" .. date .. " - " .. author .. "</p>\n" ..
            "</div></a>\n"
end

function sortStringDate(a, b)
    -- The date is in the format of YYYY-MM-DD
    local a_year = tonumber(string.sub(a, 1, 4))
    local a_month = tonumber(string.sub(a, 6, 7))
    local a_day = tonumber(string.sub(a, 9, 10))

    local b_year = tonumber(string.sub(b, 1, 4))
    local b_month = tonumber(string.sub(b, 6, 7))
    local b_day = tonumber(string.sub(b, 9, 10))

    if a_year < b_year then
        return true
    elseif a_year > b_year then
        if a_month < b_month then
            return true
        elseif a_month > b_month then
            if a_day < b_day then
                return true
            end
        end
    end
end

this:addBuilders {
    blogpostList = function()
        p = {}

        for file in io.popen('ls exampleSite/page/posts'):lines() do
            local f = fragments:getPage("posts/" .. string.sub(file, 1, -6))
            table.insert(p, {
                id = string.sub(file, 1, -6),
                title = f:getSharedMeta("postTitle"),
                date = f:getSharedMeta("postDate"),
                author = f:getSharedMeta("author"),
                description = f:getSharedMeta("postDescription")
            })
        end

        -- Sort the posts by date
        table.sort(p, function(a, b)
            return sortStringDate(a.date, b.date)
        end)

        local result = ""
        for i, v in ipairs(p) do
            result = result .. blogpost(v.id, v.title, v.date, v.author, v.description)
        end

        return result
    end
}

~~~

*{blogpostList}
