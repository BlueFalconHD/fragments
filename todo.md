
# todo

- [ ] implement hierarchy tracking
- [ ] lua embedded
  - [x] ~~first prototype using github.com/rosbit/luago~~
    - no clear way to use this effectively, using gopher-lua ~~and gopher-luar~~ will work better
    - [x] lua-utils
  - [x] introduce fragments 'library' into lua environment
    - [x] fragment type
    - [x] 'this' instance injection - partial, not integrated
    - [x] figure out what function should be pure lua vs. go
  - [ ] error handling
  - [ ] potentially add some capability to "include" lua files, could lead to more clutter though
- [ ] redo fragment parsing, currently it's a mess
  - [x] removed old rubbish, gutted main file
  - [ ] make it more modular, readable, and maintainable
- [ ] branding/logo
- [ ] better error handling overall
- [ ] documentation


## ideal usage structure
this following section includes the ways I want the project to be used

- build.sh: runs the fragments CLI command to build the site
- config.yaml: configuration file for the site
- page/: fragments that are used as pages
  - index.frag: the main page
  - other-page.frag: another page
  - potential-folder/
    - index.frag: main page for the folder
    - other-page.frag: another page in the folder
- fragment/: fragments that are used in the site
  - page.frag: the main document structure
  - footer.frag: the footer fragment
  - header.frag: the header fragment
  - breadcrumb.frag: the breadcrumb fragment
- resources/: these files will be copied to the output directory
  - css/
    - main.css: main css file
  - js/
    - main.js: main js file
  - img/
    - logo.png: the site logo

`config.yaml`:
```yaml
page_template: fragment/page.frag
```

The page template is special. Unless a fragment specifies otherwise, the set template will be used to render the page.

`fragment/page.frag`:
```html

-- we want to set some constants here for the global metadata
-- this is a fragment, so we can use the fragment functions

-- The page template also has shared metadata. The page using the template can override the values.

this.meta:shared {
    title = "Hello, World!",
    description = "This is a test site",
    stylesheet = "main.css",
    favicon = {
        32 = "favicon-32x32.png",
        16 = "favicon-16x16.png",
        svg = "favicon.svg",
    },
    appleTouchIcon = "apple-touch-icon.png",
}

===

<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>${title}</title>
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <meta name="description" content="${description}" />
  <link rel="stylesheet" type="text/css" href="${stylesheet}" />
  <link rel="icon" type="image/png" sizes="32x32" href="${favicon.32}">
  <link rel="icon" type="image/png" sizes="16x16" href="${favicon.16}">
  <link rel="apple-touch-icon" type="image/png" sizes="180x180" href="${appleTouchIcon}">
  <link rel="icon" type="image/svg+xml" href="${favicon.svg}">
</head>
<body>
    <!-- this is where the content will be inserted -->
    ${CONTENT} 
</body>
</html>
```

Any fragment can take in content as well. This can be used to make components that can be reused across the site.

`fragment/siteWrapper.frag`:
```html
-- This fragment should include the header and footer, and wrap the content in a div

===
@{fragment/header.frag}
<div class="content">
    ${CONTENT}
</div>
@{fragment/footer.frag}
```

`fragment/postCard.frag`:
```html
-- We get some content which sets the shared metadata
this.meta:shared {
    title = "Post",
    description = "A post on the site",
    date = "2021-01-01",
}

===

<div class="post-card">
    <h2>${title}</h2>
    <p>${description}</p>
    <p>${date}</p>
</div>
```

`page/index.frag`:
```html

-- This page can override the shared metadata to set the title and description
this.meta:shared {
    title = "Home",
    description = "The homepage of my site!",
}

-- Lua's facilities for dealing with the filesystem are limited,
-- so fragments' standard library includes some basic functions for reading files and directories
-- fs:readDir returns a list of files in a directory
-- fs:readFile returns the contents of a file
-- fs:writeFile writes to a file
-- fs:rmFile removes a file
-- fs:rmDir removes a directory
-- fs:mkdir creates a directory

this.builder {
  getRecentPosts = function()
    local postDir = "posts/"
    local posts = {}
    -- get every post in the posts directory
    fs:readDir(postDir):forEach(function(file)
      local post = fragment:templated(postDir .. file, "fragment/postCard.frag")
      table.insert(posts, post)
    end)
    
    -- return all of the posts joined by newlines
    return table.concat(posts, "\n")
  end,
}

===

@{fragment/siteWrapper.frag} [[
    <h1>Welcome to my site!</h1>
    <p>This is the homepage of my site.</p>
    
    <h2>Recent Posts</h2>
    <div class="recent-posts">
        *{getRecentPosts}
    </div>
]]
```

