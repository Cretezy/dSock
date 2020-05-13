# Releasing

## Prerequisites

- Bump version in `common/version.go`
- Make sure `CHANGELOG.md` is up to date and contains version changelog
- Make sure tests pass (`task tests`)

## Binaries

- Create and push version tag
- Run `task build:binaries`
- Upload all files in `build` to GitHub release (and add release notes)

## Docker

- Run `task build:docker`
- Run `task push:docker TAG=<version>` (excludes `v`)
- Run `task push:docker TAG=latest`
