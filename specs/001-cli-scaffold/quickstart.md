# Quickstart: CLI Scaffold

**Feature**: 001-cli-scaffold | **Date**: 2026-05-29

Five-minute path from a fresh clone to a working `inject` binary.

## Prerequisites

- **Go** 1.26.3+ (the version in `go.mod`)
- **make**
- **golangci-lint** (for `make lint`) — install via
  `brew install golangci-lint` or
  `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **goreleaser** (only for `make release`) — install via
  `brew install goreleaser` or see <https://goreleaser.com/install/>

## First build

```bash
git clone <repo-url>
cd injector
make help        # discover targets
make build       # produces ./bin/inject
./bin/inject     # prints help
./bin/inject project init   # prints: TODO: init!
```

Expected: every command above exits 0.

## Run the tests

```bash
make test
```

Expected: all tests pass, suite completes in < 30s, no network calls.

## Run the linter

```bash
make lint
```

Expected: zero issues on a clean checkout.

## Ctrl-C behavior

```bash
./bin/inject project init
# (this slice exits before you can interrupt it; once long-running
#  subcommands exist, Ctrl-C will cancel them and the process will exit 130)
```

## CI

Open a PR against `main`. The `CI` workflow runs `make test` and
`make lint` on `ubuntu-latest`; the result appears as a check on the PR.

## Cut a release (maintainers)

```bash
git tag v0.1.0
git push origin v0.1.0
make release     # requires GITHUB_TOKEN with repo:write
```

Expected: `darwin/{amd64,arm64}` and `linux/{amd64,arm64}` archives are
attached to the GitHub Release for `v0.1.0`.
