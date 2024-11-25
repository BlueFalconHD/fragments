this:template("post")

this:builders {
    renderMarkdown = function(content)
        -- Replace this with your own markdown rendering code
        -- For example, we will reverse the content
        return string.reverse(content)
    end
}

---

THIS IS THE MARKDOWN TEMPLATE WOO HOO
=====================================

*{renderMarkdown[[${CONTENT}]]}

=====================================