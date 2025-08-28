this:setTemplate("post")

this:setSharedMeta {
    postTitle = "Announcing Fragments 1.0",
    postDescription = "Stable builders and templates, improved configuration, and a richer example site.",
    postDate = "2024-12-01",
    author = "Hayes"
}

~~~

## Fragments 1.0 is here

Fragments is a focused, text‑first site system that keeps content maintainable while giving you programmable power when you need it.

Everything is a fragment. Compose small pieces, render Markdown, and use Lua to orchestrate your site.

### What’s new

- Stable template semantics
  - A dedicated page template with a clear content slot
  - A post template that renders Markdown with GFM features (tables, task lists, footnotes)
- Cleaner configuration and path handling
  - Uses your config’s site root and path joins under the hood
- Cross‑platform page discovery for blog lists
  - Builders can list pages without shelling out
  - Example: list posts using the fragments module
- Better date handling for listings
  - Simple ISO date comparison (newest‑first) and nicer display formatting
- Richer example site
  - Home, Blog, About, and several Posts
  - Reusable nav and footer fragments
  - Optional styles fragment for inline overrides

### Upgrading at a glance

- Prefer ISO dates (YYYY‑MM‑DD) in post metadata, for example:
```
postDate = "2024-12-01"
```

- Replace shell-based listing with the fragments module, for example:
```
for id, f in pairs(fragments:getPagesUnder("posts")) do
    -- fetch meta and build listing
end
```

- Keep reusable UI and site chrome in fragments
  - page handles document structure and layout
  - nav contains the header and links
  - footer contains attribution and links
  - post wraps article content and renders it as Markdown

### Quick example

A minimal post typically:
1. Sets the post template:
```
this:setTemplate("post")
```
2. Provides shared metadata (title, description, date, author) above the content separator.
3. Writes Markdown content below the separator.

A simple page that lists recent posts:
1. Sets the page template:
```
this:setTemplate("page")
```
2. Provides a shared title in the Lua section.
3. Adds a heading and includes a listing of posts in the content body.

### Feature highlights

- Markdown rendering supports:
  - Task lists
  - Tables
  - Footnotes
  - Strikethrough
- Builder functions return strings you can inject anywhere in the content
- Shared and local metadata give you flexible scoping

Example table:

| Feature        | Status |
|----------------|--------|
| Templates      | Stable |
| Markdown (GFM) | Stable |
| Builders       | Stable |

Footnote example.[^fn1]

[^fn1]: Rendered with a Markdown engine that supports footnotes.

### Roadmap

- Developer‑friendly CLI (init and build)
- Asset pipeline improvements
- More fragment utilities and patterns

Thanks for trying Fragments and building with it. I’m excited to see what you create.

— Hayes
