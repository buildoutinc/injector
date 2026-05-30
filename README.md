# injector

`inject` is a CLI for managing secrets and configuration across an
organization's projects, services, and environments. See
[`# Injector Project Constitution`](./%23%20Injector%20Project%20Constitution)
for the governing principles.

## Quick start

```bash
make help          # list every Makefile target
make build         # builds ./bin/inject
./bin/inject       # prints help
./bin/inject project init    # → "TODO: init!"
make test          # run the Go test suite (offline)
make lint          # run golangci-lint
```

## Shell completion (optional)

```bash
# bash, zsh, fish, elvish, powershell, nushell, xonsh, tcsh
source <(./bin/inject _carapace bash)
```

## Releasing

Tag the commit and run `make release` (requires `goreleaser` and a
`GITHUB_TOKEN` with `repo:write`). Artifacts are built for
`darwin/{amd64,arm64}` and `linux/{amd64,arm64}` and published to the
matching GitHub Release.

## Continuous integration

`.github/workflows/ci.yml` runs `make test` and `make lint` on every PR
against `main` and every push to `main`.
