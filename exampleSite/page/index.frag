-- lua code segment
this:setSharedMeta("title", "Home")
this:setSharedMeta("description", "This is the home page")

function getFormattedDate()
  return os.date("%Y-%m-%d")
end

this:setMeta("date", getFormattedDate())
===

${header}

<h1 class="title">Welcome to my site!</h1>
<h2 class="description">This is my home page.</h2>

${blogposts}

${footer}
