# Contributing

Thanks for contributing to lexkey.

## Setup

1. Fork and clone the repository.
2. `go mod download`
3. `go test ./...`

## Pull requests

- Run `go fmt ./...` and `go vet ./...`.
- Any wire-format change requires [SPEC.md](SPEC.md) updates and tests for sort order.
- Update `docs/` for new types or encoding rules.

## Changelog

Note changes under `## [Unreleased]` in [CHANGELOG.md](CHANGELOG.md).
