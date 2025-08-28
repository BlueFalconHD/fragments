this:addBuilders {
    siteTitle = function()
        -- Allow pages to override the title via shared or local meta; fallback to a default
        local title = this:getSharedMeta("title")
        if title == nil or tostring(title) == "nil" or tostring(title) == "" then
            title = this:getLocalMeta("title")
        end
        if title == nil or tostring(title) == "nil" or tostring(title) == "" then
            title = "Fragments Site"
        end
        return tostring(title)
    end
}

~~~
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>*{siteTitle}</title>
  @{sitemeta}
  <link rel="stylesheet" href="/style.css">
  @{styles}
</head>
<body>
  @{nav}

  <main>
    ${CONTENT}
  </main>

  @{footer}
</body>
</html>
