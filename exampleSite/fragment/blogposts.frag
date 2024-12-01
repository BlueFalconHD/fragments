

function blogpost(path)
    print("Path: " .. path)
    f = getFragment(this, path)
    print("Fragment: " .. f.code)
    return string.format("<div class='blogpost'><h2>%s</h2><h3>%s on %s</h3><p>%s</p></div>", f:getMeta("postTitle"), f:getMeta("author"), f:getMeta("postDate"), f:getMeta("postDescription"))
end

this:builders {
    blogpostList = function(content)
        -- For now let's just say there is one blogpost, 'posts/example.frag'
        p = { "../page/posts/example" }


        local result = ""
        for i, v in ipairs(p) do
            result = result .. blogpost(v)
        end

        return result
    end
}


---

<h1>Blogposts</h1>
*{blogpostList}