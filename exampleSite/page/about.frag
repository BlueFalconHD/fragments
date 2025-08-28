this:setTemplate("page")

this:setSharedMeta {
    title = "About",
    funFact = "I once built a whole site from scratch in a weekend."
}

~~~

@{markdown[[# About

This site is built with Fragments, a static site generator that uses Lua for templating and content generation.

]]}

@{markdown[[## Recent posts]]}
@{blogposts}
