this:setTemplate("post")

this:setSharedMeta {
    postTitle = "Advanced Composition with Fragments",
    postDescription = "Patterns for slots, nested fragments, and builder pipelines.",
    postDate = "2024-07-12",
    author = "Hayes"
}

this:addBuilders {
    callout = function(content)
        local html = renderMarkdown(content)
        if html == nil then
            html = content
        end
        return "<div style='margin:16px 0;padding:14px 16px;border-left:4px solid #89b4fa;background:#181825;border-radius:6px;color:#cdd6f4'>" .. html .. "</div>"
    end,
    twoCol = function(content)
        local sepStart, sepEnd = string.find(content, "\n---\n", 1, true)
        local left = content
        local right = ""
        if sepStart then
            left = string.sub(content, 1, sepStart - 1)
            right = string.sub(content, sepEnd + 1)
        end

        local leftHTML = renderMarkdown(left)
        if leftHTML == nil then leftHTML = left end
        local rightHTML = renderMarkdown(right)
        if rightHTML == nil then rightHTML = right end

        local wrapOpen = "<div style='display:grid;gap:16px;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));margin:16px 0'>"
        local colStyle = "padding:12px;border:1px solid #313244;border-radius:8px;background:#1e1e2e"
        local leftCol = "<div style='" .. colStyle .. "'>" .. leftHTML .. "</div>"
        local rightCol = "<div style='" .. colStyle .. "'>" .. rightHTML .. "</div>"
        local wrapClose = "</div>"
        return wrapOpen .. leftCol .. rightCol .. wrapClose
    end,
    linkButton = function(content)
        local bar = string.find(content, "|", 1, true)
        if not bar then
            return content
        end
        local href = string.sub(content, 1, bar - 1)
        local label = string.sub(content, bar + 1)
        local style = "display:inline-block;background:#89b4fa;color:#11111b;padding:8px 14px;border-radius:8px;text-decoration:none;font-weight:600"
        return "<a href='" .. href .. "' style='" .. style .. "'>" .. label .. "</a>"
    end
}

~~~

# Advanced composition with Fragments

Fragments lets you mix declarative content with programmable building blocks.
This post explores practical patterns for composing pages from small parts.

## Slots with nested fragments

Use fragments as slots by passing content to them. The `ihavecontent` fragment
accepts content and renders it inline.

@{ihavecontent[[This content is passed into a fragment as a “slot.” You can
think of it like a component with a default body that you can override.]]}

## Builder pipelines inside Markdown

You can compute HTML from inside your Markdown using builders. The result
is injected before Markdown is rendered by the post template.

*{callout[[Pro tip: Builders are just Lua functions that return strings.
They’re perfect for small, reusable UI patterns.]]}

Here’s another example that renders two columns from a single content block:

*{twoCol[[
### Left column
- Write Markdown
- Add lists and links
- Keep content focused

---

### Right column
- Use grids with inline styles
- Compose small, reusable units
- Keep layout minimal
]]}

### Notes
This separator (`---`) splits the content into two columns. Everything above it
is the left column; everything below it becomes the right column.

## Cross-linking with styled buttons

Use a tiny builder to create styled links:

*{linkButton[[/about.html|Learn more about this site]]}

## Wrap-up

- Compose pages from small fragments
- Use builders for repeatable UI patterns
- Pass content into fragments when you need slots
- Keep Markdown as your main authoring format

Fragments stays out of your way when you want static content—and gives you power
when you need programmability.
