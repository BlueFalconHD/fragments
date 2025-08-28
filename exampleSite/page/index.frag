this:setTemplate("page")

this:setSharedMeta {
    title = "Home"
}

~~~

<h1 class="title">Welcome to my site!</h1>
<p class="description">This is a simple site built with <a href="https://github.com/bluefalconhd/fragments">Fragments</a>.</p>
<p class="description">Latest articles below:</p>

<h1>Recent Blogposts</h1>
@{blogposts}
