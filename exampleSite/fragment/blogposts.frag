function formatDate(d)
    if d == nil or d == "" then return "" end
    local y = string.sub(d, 1, 4)
    local m = tonumber(string.sub(d, 6, 7))
    local dayNum = tonumber(string.sub(d, 9, 10))
    local MONTHS = {"Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"}
    local mname = MONTHS[m] or string.sub(d, 6, 7)
    return mname .. " " .. tostring(dayNum) .. ", " .. y
end

function blogpost(id, title, date, author, description)
    local displayDate = formatDate(date)
    local dateHtml = ""
    if displayDate ~= "" then
        dateHtml = " <i class='secondary'>(" .. displayDate .. ")</i>"
    end
    return  "<a class='unstyled-link' href='posts/" .. id .. ".html'><div class='blogpost'>\n" ..
            "   <h3>" .. title .. dateHtml .. "</h3>\n" ..
            "   <p>" .. description .. "</p>\n" ..
            "</div></a>\n"
end

function sortStringDate(a, b)
    -- Expect YYYY-MM-DD; return true if a is more recent than b (newest-first)
    if a == nil and b == nil then return false end
    if a == nil then return false end
    if b == nil then return true end
    if a == b then return false end
    -- Lexicographic compare works for ISO dates; invert for descending
    return a > b
end




this:addBuilders {
    blogpostList = function()
        p = {}

        for id, f in pairs(fragments:getPagesUnder("posts")) do
            local title = f:getSharedMeta("postTitle") or f:getLocalMeta("postTitle") or id
            local date = f:getSharedMeta("postDate") or f:getLocalMeta("postDate") or ""
            local author = f:getSharedMeta("author") or f:getLocalMeta("author") or ""
            local description = f:getSharedMeta("postDescription") or f:getLocalMeta("postDescription") or ""
            table.insert(p, {
                id = id,
                title = title,
                date = date,
                author = author,
                description = description
            })
        end

        -- Sort the posts by date
        table.sort(p, function(a, b)
            return sortStringDate(a.date, b.date)
        end)

        if #p == 0 then
            return "<p class='description'>No posts yet.</p>"
        end

        local result = ""
        for i, v in ipairs(p) do
            result = result .. blogpost(v.id, v.title, v.date, v.author, v.description)
        end

        return result
    end
}

~~~

*{blogpostList}
