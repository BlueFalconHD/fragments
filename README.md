
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

this:setSharedMeta {
    buildDate = getStringFormattedDate(),
    title = "Hello, World!"
}

this:addBuilders {
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

~~~

The content of our fragment begins here.

By using a dollar sign and braces, you can include metadata set in the lua environment: ${buildDate}

To include other fragments, you can use @{fragmentName}

Finally, you can dynamically run a lua function that returns a string, like so: *{randomBuilder}
```

### CLI

Use the CLI to initialize a new project and build your site.

Initialize a new project (creates config.yml, fragment/page/include folders, and example files):

```
fragments init mysite
```

Build the site (outputs to the configured build directory and copies include assets):

```
fragments build -c path/to/config.yml
```

Common example when run from the project root:

```
fragments build -c config.yml
```

### CI

Continuous Integration runs on pushes and pull requests to the `main` branch across Linux, macOS, and Windows. It installs Go 1.19.x, downloads dependencies, vets, builds, and tests (race on non‑Windows).

- Workflow file: `.github/workflows/ci.yml`
- What it does:
  - `go mod download`
  - `go vet ./...`
  - `go build ./...`
  - `go test -race ./...` (Linux/macOS) or `go test ./...` (Windows)

### Releases

Releases are automated with GoReleaser on tag pushes. Tag the repository and push the tag to trigger a release build.

```
git tag v0.1.0
git push origin v0.1.0
```

- Workflow file: `.github/workflows/release.yml`
- GoReleaser config: `.goreleaser.yaml`
- Outputs:
  - Cross‑platform archives for Linux, macOS, and Windows (amd64/arm64)
  - SHA256 checksums
  - GitHub Release with attached artifacts

Optional local snapshot (no publish):

```
goreleaser release --snapshot --clean --config fragments/.goreleaser.yaml
```