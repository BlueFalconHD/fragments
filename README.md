**project direction**

I have picked this project up again and intend on polishing it into a usable product. As of right now I am working on an embedded lua scripting API for more dynamic generation of pages, rather than the statically compiled scripts I had used initially. This makes the project a lot more complex though, which is probably why it won't be in a usable state for a few more weeks. I intend to finish this project by the end of High Seas, an event by Hack Club. Switch to the [docs-todo-and-ideal-project](https://github.com/BlueFalconHD/fragments/tree/docs-todo-and-ideal-project) (horrible branch name idk what I was doing) branch and view `todo.md` to see how I want the project to be used. I would love this to be usable by everyone so any suggestions are welcome (submit an issue).

---

everything is a fragment

work in progress, currently fragment parsing and evaluating works.

a fragment defines any meta at the top of the file:

```
---
key: value
---
```

then it defines whatever it wants after that:

```
anything ${i} @{want}
```

- `${}` insert meta value
- `@{}` insert another fragment


ex.


```
---
title: Home/Welcome
siteName: Example Site
---
Welcome to ${siteName}.
Today's date is ${date}.
Test undefined meta: ${undefined}

@{footer}
```
