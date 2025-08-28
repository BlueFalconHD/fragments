# Roadmap

A focused, categorized plan for upcoming work. This replaces earlier completed items with concrete, actionable tasks.

## CLI & Dev Workflow
- [ ] Watch mode and dev server
  - [ ] Rebuild on file changes for fragment/page/include/config
  - [ ] Optional local static server (serves build/) with live reload
  - [ ] Cross‑platform file watching (fsnotify or similar)
- [ ] Concurrency controls
  - [ ] Parallel page builds via a worker pool
  - [ ] -j/--jobs flag to control parallelism
  - [ ] Ensure thread‑safety for cache and evaluation
- [ ] Logging flags
  - [ ] -v/--verbose for debug output
  - [ ] --quiet to suppress non‑errors
  - [ ] Consistent, structured logs
- [ ] Pretty/minify toggle
  - [ ] --pretty (default) and --minify output options

## Configuration & Validation
- [ ] Config validation
  - [ ] Validate required fields and paths
  - [ ] Friendly error messages with suggestions
  - [ ] Reasonable defaults for missing fields
- [ ] Config discovery
  - [ ] Search up from CWD for config.yml if not provided
  - [ ] Environment variable override for config location

## Core Engine & Semantics
- [ ] Cycle detection for fragment references
  - [ ] Detect recursive includes and show a clear include stack
  - [ ] Configurable limit for maximum include depth
- [ ] Named slots
  - [ ] Multiple content slots per fragment/template (e.g., header, content, footer)
  - [ ] Backward compatible default slot behavior
- [ ] Standard library of builders
  - [ ] date.format, date.parse (simple ISO helpers)
  - [ ] slugify
  - [ ] url.join (safe join of base and path)
  - [ ] list.sort (stable comparator options)
  - [ ] html.escape / unescape
  - [ ] string utilities commonly needed in content
- [ ] Safer Markdown option
  - [ ] renderMarkdownSafe with unsafe HTML disabled
  - [ ] Config or builder‑level toggle to select safe vs. unsafe rendering
- [ ] Enhanced Lua error surfacing
  - [ ] Include builder/fragment name, include stack, and source snippet
  - [ ] Consistent formatting for parse vs. runtime errors
  - [ ] Suggestions when a builder or fragment is missing

## Performance & Caching
- [ ] AST caching and mtime invalidation
  - [ ] Cache parsed ASTs keyed by absolute path + mtime
  - [ ] Invalidate affected dependents when a fragment changes
- [ ] Evaluation memoization
  - [ ] Skip re‑evaluation when inputs (meta, template, content) are unchanged
  - [ ] Clear cache on config or stdlib changes
- [ ] Optional LState pooling
  - [ ] Reuse Lua states in a safe, isolated way to reduce allocations
  - [ ] Document boundaries and caveats

## Output & SEO
- [ ] Sitemap.xml
  - [ ] Generate from discovered pages with configurable base URL
  - [ ] Exclude patterns (e.g., drafts) via config
- [ ] RSS/Atom/JSON Feed
  - [ ] Feed for posts under a configured prefix (e.g., posts)
  - [ ] Include title, description, date, and canonical link
  - [ ] Limit and sort options in config

## Cross‑Platform & Paths
- [ ] Path normalization
  - [ ] Centralize path joins and normalization (Windows/macOS/Linux)
  - [ ] Use forward‑slash internal representation; convert at boundaries
- [ ] File discovery robustness
  - [ ] Respect case sensitivity differences
  - [ ] Guard against symlink cycles in discovery

## Quality & Reliability
- [ ] Tests
  - [ ] Lexer unit tests (balanced braces/brackets, escapes)
  - [ ] Parser unit tests (nodes, nested content, errors)
  - [ ] Evaluator unit tests (meta resolution, builders, fragments)
  - [ ] Golden tests for error messages and formatted snippets
  - [ ] Integration tests building the example site and snapshotting output
- [ ] Property/fuzz tests
  - [ ] Fuzz parser and content blocks for panic resistance
- [ ] Determinism
  - [ ] Ensure stable output ordering for listings and maps

## Packaging & Releases
- [ ] Package layout
  - [ ] Move CLI to cmd/fragments
  - [ ] Split engine/runtime/parser into internal packages
  - [ ] Keep public API minimal and documented
- [ ] CI
  - [ ] GitHub Actions for build, vet, test, race
  - [ ] Cache dependencies for speed
- [ ] Releases
  - [ ] Goreleaser for multi‑platform binaries (darwin/arm64, linux/amd64, windows/amd64)
  - [ ] Checksums and signed releases

## Editor & DX
- [ ] Editor tooling
  - [ ] VS Code syntax highlighting for .frag (meta, builders, fragments, slots)
  - [ ] Optional Tree‑sitter grammar
  - [ ] Lightweight linter for common mistakes (unbalanced delimiters, unknown references)
- [ ] Templates and snippets
  - [ ] Common fragment snippets (page, post, nav, footer)
  - [ ] Builder function snippet templates

## Documentation
- [ ] Quickstart
  - [ ] Install, init, build, dev (watch/server)
- [ ] CLI reference
  - [ ] init, build, watch, flags (config, jobs, verbose, quiet, minify)
- [ ] Fragment syntax
  - [ ] Meta, builder references, fragment references, named slots
- [ ] Lua API and stdlib builders
  - [ ] this, fragments module, stdlib functions with examples
- [ ] Templates and composition patterns
  - [ ] Page/post templates, slots, nested fragments
- [ ] SEO helpers
  - [ ] sitemeta, head meta, canonical URLs, JSON‑LD usage
- [ ] Troubleshooting and error guide
  - [ ] Common errors, stack traces, how to debug
- [ ] Upgrade guide
  - [ ] Breaking changes and migration steps
- [ ] Example site walkthrough
  - [ ] How it’s structured; copying patterns to your site
- [ ] Contribution guide and architecture overview
  - [ ] How to build, test, and propose changes