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
