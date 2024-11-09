
# fragments
## everything is a fragment

work in progress

### what is a fragment?

A fragment is the basic building block of your site. When I say 'everything is a fragment', I mean that every piece of content on your site is a fragment. Fragments can be...

- Pages
- Components
- Templates
- ...

What is so powerful about fragments compared to other SSGs is that each fragment has access to a lua environment, which can be used to generate meta, content, and more.

### how do I use fragments?

Fragments are defined in a `fragments` directory in your project (this is configurable). Each fragment is a file with a .frag extension.

```lua
-- fragments/hello.frag
-- this is the lua body of our fragment

-- we can do some cool stuff here

function getStringFormattedDate()
    return os.date("%Y-%m-%d")
end

this:meta {
    buildDate = getStringFormattedDate(),
    title = "Hello, World!"
}

this:builders {
    randomBuilder = function()
        -- Pick 40 random characters from the alphabet and append them to a string, then return it
        local alphabet = "abcdefghijklmnopqrstuvwxyz"
        local result = ""
        for i = 1, 40 do
            result = result .. alphabet:sub(math.random(1, #alphabet), math.random(1, #alphabet))
        end
        return result
    end
}

---

The content of our fragment begins here.

By using a dollar sign and braces, you can include metadata set in the lua environment: ${buildDate}

To include other fragments, you can use @{fragmentName}

Finally, you can dynamically run a lua function that returns a string, like so: *{randomBuilder}
```