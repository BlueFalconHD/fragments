# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

## [Unreleased]

### Added
- CLI entrypoint with `init` and `build` subcommands.
- GitHub Actions CI workflow for build/vet/test across Linux, macOS, and Windows.
- Goreleaser configuration for cross‑platform builds and checksums.
- Basic CHANGELOG and README notes for CLI usage.

### Changed
- Support dotted meta keys in content by resolving nested meta paths.

### Fixed
- Nil dereference risks in `FragmentCache.Add` and when assigning `${CONTENT}` to templates.

## [0.1.0] - YYYY-MM-DD
- Initial tagged release.

[Unreleased]: https://github.com/bluefalconhd/fragments/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/bluefalconhd/fragments/releases/tag/v0.1.0