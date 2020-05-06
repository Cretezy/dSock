# Releasing

## Prerequisites

- Bump version in `common/version.go`
- Make sure `CHANGELOG.md` is up to date and contains version changelog
- Make sure tests pass (`scripts/run-tests`)

## Binaries

- Create and push version tag
- Run `scripts/build-binaries`
- Upload all files in `build` to GitHub release

## Docker

- Run `scripts/build-docker`
- Run `scripts/push-images <version>` (excludes `v`)
- Run `scripts/push-images latest`
