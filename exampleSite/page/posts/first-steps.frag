this:setTemplate("post")

this:setSharedMeta {
    postTitle = "First Steps with Fragments",
    postDescription = "A guided tour: build pages, compose fragments, and sprinkle in dynamic content.",
    postDate = "2024-01-05",
    author = "Hayes"
}

~~~

# First steps with Fragments

Fragments makes it easy to build a site from small, composable parts. In this post you’ll:
- Create a page backed by a template
- Reuse fragments like a site nav and footer
- Render Markdown with GFM features (tables, task lists, footnotes)
- Understand builders and shared metadata

> The best systems stay out of your way—until you need power.

## Quick start

Create a new page fragment. Set the template, then write your content under the `~~~` separator.

    this:setTemplate("page")
    this:setSharedMeta {
        title = "Hello World"
    }

    ~~~

    # Hello World
    This page is powered by a template. Neat!

The template typically includes the site header, footer, and a `\${CONTENT}` placeholder where your page body is injected.

## Composition

Fragments compose like LEGO bricks. For example:
- A `page` template composes `nav` and `footer`
- A `post` template composes `markdown` rendering
- Your content composes within the template’s `\${CONTENT}` slot

Here’s a conceptual map:

| Fragment     | Purpose                        |
|--------------|--------------------------------|
| `page`       | Base HTML document + layout    |
| `nav`        | Top navigation bar             |
| `footer`     | Footer with attribution        |
| `post`       | Article wrapper + Markdown     |
| `blogposts`  | Lists posts with meta          |

## Dynamic content with builders

Builders are Lua functions you can call from your content to compute strings. Inside the `post` template, content is rendered via a `markdown` builder that supports GFM.

Example task list:

- [x] Set a page template
- [x] Write your first post
- [ ] Add search
- [ ] Ship it

Footnotes are supported too.[^first]

[^first]: You’re reading a footnote rendered by Goldmark with the Footnote extension.

## A tiny example post fragment

Below is a small example post showing the common metadata set for a blog post:

    this:setTemplate("post")

    this:setSharedMeta {
        postTitle = "My Tiny Post",
        postDescription = "A short example post to test the listing.",
        postDate = "2023-11-20",
        author = "Hayes"
    }

    ~~~

    # My Tiny Post
    This is just a few lines of content to verify listings and sorting.

When you add multiple posts, the blog listing will sort them by `postDate` (newest first).

## Tips

- Prefer ISO dates (`YYYY-MM-DD`) for easy sorting.
- Keep fragments small and focused.
- Push shared values (like site-wide title) into shared meta so templates can access them.

## What’s next

- Add two more posts with different `postDate` values to see sorting in action
- Customize the `nav` and `footer` fragments
- Try adding a builder that generates a tag cloud from post metadata

Happy building!
