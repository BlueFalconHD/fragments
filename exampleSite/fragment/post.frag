this:setTemplate("page")

-- We have the following meta available to us for a standard post
-- postTitle
-- postDate
-- postDescription
-- author

---

<h1>${postTitle}</h1>
<p>${postDescription}</p>
<p>${postDate} - ${author}</p>

=================================

@{markdown[[${CONTENT}]]}

=================================