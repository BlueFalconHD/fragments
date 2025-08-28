this:setSharedMeta {
    site = {
        name = "My Site",
        title = "My Site",
        description = "Just another site built with Fragments.",
        author = "Your Name",
        baseUrl = "",

        -- Optional social handles/links
        twitter = "",
        github = "",

        -- Optional nav model (not used directly by this fragment, but handy to centralize)
        nav = {
            { href = "/index.html", label = "Home" },
            { href = "/blog.html",  label = "Blog" },
            { href = "/about.html", label = "About" }
        }
    }
}

this:addBuilders {
    -- Returns a computed full page title: "<Page or Post Title> — <Site Title>"
    computedTitle = function()
        local function toStr(v)
            if v == nil or tostring(v) == "nil" then return "" end
            return tostring(v)
        end

        local siteTitle = toStr(this:getSharedMeta("site.title"))
        if siteTitle == "" then siteTitle = "Fragments Site" end

        local pageTitle = toStr(this:getSharedMeta("title"))
        if pageTitle == "" then
            pageTitle = toStr(this:getSharedMeta("postTitle"))
        end

        if pageTitle == "" then
            return siteTitle
        end
        return pageTitle .. " — " .. siteTitle
    end,

    -- Returns a canonical URL using site.baseUrl and an optional path passed as content.
    -- Example usage: *{canonical[[/about.html]]}
    canonical = function(content)
        local function toStr(v)
            if v == nil or tostring(v) == "nil" then return "" end
            return tostring(v)
        end
        local base = toStr(this:getSharedMeta("site.baseUrl"))
        local path = toStr(content)
        if base == "" then
            return path
        end
        if path == "" then
            return base
        end
        -- Ensure single slash joining
        if string.sub(base, -1) == "/" and string.sub(path, 1, 1) == "/" then
            return base .. string.sub(path, 2)
        elseif string.sub(base, -1) ~= "/" and string.sub(path, 1, 1) ~= "/" then
            return base .. "/" .. path
        else
            return base .. path
        end
    end,

    -- Emit common head meta tags. Optionally pass a path (e.g. "/posts/hello.html")
    -- to set canonical and og:url correctly: *{headMeta[[/posts/hello.html]]}
    headMeta = function(content)
        local function toStr(v)
            if v == nil or tostring(v) == "nil" then return "" end
            return tostring(v)
        end

        local siteName = toStr(this:getSharedMeta("site.name"))
        if siteName == "" then siteName = toStr(this:getSharedMeta("site.title")) end
        if siteName == "" then siteName = "Fragments Site" end

        local author = toStr(this:getSharedMeta("author"))
        if author == "" then author = toStr(this:getSharedMeta("site.author")) end

        local desc = toStr(this:getSharedMeta("postDescription"))
        if desc == "" then desc = toStr(this:getSharedMeta("site.description")) end

        -- Prefer explicit page/post title, else fallback to computed site title
        local title = toStr(this:getSharedMeta("title"))
        if title == "" then title = toStr(this:getSharedMeta("postTitle")) end
        if title == "" then title = siteName end

        local canonicalUrl = this:builders().canonical(content)

        local twitter = toStr(this:getSharedMeta("site.twitter"))
        -- Normalize twitter handle to @name if it isn't already
        if twitter ~= "" and string.sub(twitter, 1, 1) ~= "@" then
            twitter = "@" .. twitter
        end

        local parts = {}
        local function push(s) table.insert(parts, s) end

        if desc ~= "" then
            push('<meta name="description" content="' .. desc .. '">')
        end
        if author ~= "" then
            push('<meta name="author" content="' .. author .. '">')
        end

        -- Open Graph
        push('<meta property="og:type" content="website">')
        push('<meta property="og:site_name" content="' .. siteName .. '">')
        push('<meta property="og:title" content="' .. title .. '">')
        if desc ~= "" then
            push('<meta property="og:description" content="' .. desc .. '">')
        end
        if canonicalUrl ~= "" then
            push('<meta property="og:url" content="' .. canonicalUrl .. '">')
            push('<link rel="canonical" href="' .. canonicalUrl .. '">')
        end

        -- Twitter
        push('<meta name="twitter:card" content="summary">')
        if twitter ~= "" then
            push('<meta name="twitter:site" content="' .. twitter .. '">')
            push('<meta name="twitter:creator" content="' .. twitter .. '">')
        end
        push('<meta name="twitter:title" content="' .. title .. '">')
        if desc ~= "" then
            push('<meta name="twitter:description" content="' .. desc .. '">')
        end

        return table.concat(parts, "\n")
    end,

    -- Emit a small JSON-LD script for either a blog post (if postTitle present) or the site.
    -- Usage: *{jsonLd[[/posts/hello.html]]} to include canonical URL in @id/url fields.
    jsonLd = function(content)
        local function toStr(v)
            if v == nil or tostring(v) == "nil" then return "" end
            return tostring(v)
        end

        local isPost = this:getSharedMeta("postTitle") ~= nil and tostring(this:getSharedMeta("postTitle")) ~= "nil"
        local canonicalUrl = this:builders().canonical(content)
        local siteName = toStr(this:getSharedMeta("site.name"))
        if siteName == "" then siteName = toStr(this:getSharedMeta("site.title")) end
        if siteName == "" then siteName = "Fragments Site" end

        if isPost then
            local title = toStr(this:getSharedMeta("postTitle"))
            local desc = toStr(this:getSharedMeta("postDescription"))
            local date = toStr(this:getSharedMeta("postDate"))
            local author = toStr(this:getSharedMeta("author"))
            local json = '{'
                .. '"@context":"https://schema.org",'
                .. '"@type":"BlogPosting",'
                .. '"headline":"' .. title .. '",'
                .. (desc ~= "" and ('"description":"' .. desc .. '",') or "")
                .. (date ~= "" and ('"datePublished":"' .. date .. '",') or "")
                .. (author ~= "" and ('"author":{"@type":"Person","name":"' .. author .. '"},') or "")
                .. (canonicalUrl ~= "" and ('"mainEntityOfPage":{"@type":"WebPage","@id":"' .. canonicalUrl .. '"},') or "")
                .. '"publisher":{"@type":"Organization","name":"' .. siteName .. '"}'
                .. '}'
            return '<script type="application/ld+json">' .. json .. '</script>'
        else
            local base = toStr(this:getSharedMeta("site.baseUrl"))
            local json = '{'
                .. '"@context":"https://schema.org",'
                .. '"@type":"WebSite",'
                .. '"name":"' .. siteName .. '",'
                .. (base ~= "" and ('"url":"' .. base .. '"') or ('"url":"' .. canonicalUrl .. '"'))
                .. '}'
            return '<script type="application/ld+json">' .. json .. '</script>'
        end
    end
}

~~~
*{headMeta}
*{jsonLd}
