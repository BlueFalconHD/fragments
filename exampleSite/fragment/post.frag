this:setTemplate("page")

-- We have the following meta available to us for a standard post
-- postTitle
-- postDate
-- postDescription
-- author

~~~
<article>
    <h1>${postTitle}</h1>
    @{markdown[[${CONTENT}]]}
</article>