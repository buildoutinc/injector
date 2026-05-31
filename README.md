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
./bin/inject version         # prints the running version
make test          # run the Go test suite (offline)
make lint          # run golangci-lint
```

## Shell completion (optional)

```bash
# bash, zsh, fish, elvish, powershell, nushell, xonsh, tcsh
source <(./bin/inject _carapace bash)
```

## Upgrading

`inject` ships with a built-in self-update flow that pulls the latest
release directly from GitHub.

```bash
inject upgrade --check    # report whether a newer release exists; exit 0
inject upgrade            # download, verify, and swap the binary
inject upgrade --pre-release   # include pre-release tags
```

Behavior:

- Compares the running binary's semver to the latest GitHub Release for
  `buildoutinc/injector`.
- Verifies the SHA-256 checksum of the downloaded archive against
  `checksums.txt` before swapping anything.
- Refuses (and tells you to use your package manager instead) when the
  binary path isn't writable by your user — e.g., the Homebrew cellar.

### Passive update notice

When you run `inject`, `inject help`, `inject --help`, or `inject
version`, the CLI checks GitHub at most **once per 24 hours** and
appends a single-line nudge when a newer version is available:

```text
inject v0.1.0 (commit abc1234, built 2026-05-30)

==> A newer version is available: v0.3.0
==> Run "inject upgrade" to install it.
```

The check is asynchronous and capped at 200 ms — it never slows down a
foreground command, and a network failure is silently ignored.

The cache lives under your user cache directory:

- Linux: `${XDG_CACHE_HOME:-~/.cache}/inject/update-check.json`
- macOS: `~/Library/Caches/inject/update-check.json`

### Opting out

Set `INJECT_NO_UPDATE_CHECK=1` (any non-empty value). This disables the
passive check entirely — no GitHub requests, no notice, no cache writes.
The `inject upgrade` command itself still works because it is an
explicit user action.

## Releasing

Releases are cut by pushing a semver-shaped tag to `main`. A GitHub
Actions workflow (`.github/workflows/release.yml`) handles the rest.

### Commit convention

Use **Conventional Commits** (`feat:`, `fix:`, `perf:`, `refactor:`,
`docs:`, `chore:`, `ci:`, `build:`, `test:`). The changelog is grouped
by type; any commit without a recognized prefix appears under
"Other Changes" (nothing is silently dropped).

### Cutting a release

```bash
git checkout main
git pull
# commits with Conventional Commit prefixes already on main
git tag -a v0.3.0 -m "v0.3.0"
git push origin v0.3.0
```

The workflow then:

1. **Validates** the tag matches `vX.Y.Z[-prerelease][+build]`; fails
   fast on a malformed tag.
2. **Generates** release notes with `git-chglog` using the previous tag
   as the lower bound.
3. **Builds** the binary for `linux/{amd64,arm64}` and
   `darwin/{amd64,arm64}` via GoReleaser, embedding the tag version
   into the binary at link time.
4. **Uploads** the four archives plus `checksums.txt` to the GitHub
   Release for the tag.

### Release artifacts

Each release publishes these assets:

| Filename | Contents |
|----------|----------|
| `inject_<X.Y.Z>_linux_amd64.tar.gz` | Linux x86_64 binary archive |
| `inject_<X.Y.Z>_linux_arm64.tar.gz` | Linux arm64 binary archive |
| `inject_<X.Y.Z>_darwin_amd64.tar.gz` | macOS x86_64 binary archive |
| `inject_<X.Y.Z>_darwin_arm64.tar.gz` | macOS Apple Silicon binary archive |
| `checksums.txt` | SHA-256 of each archive (GoReleaser default format) |

Inside each archive: a single executable named `inject` at the root,
matching the naming pattern the self-update library expects so
`inject upgrade` can locate and install it automatically.

### Cutting a release locally (maintainers)

```bash
make notes       # render the changelog for the latest tag to /tmp/notes.md
make release     # goreleaser release --clean --release-notes /tmp/notes.md
```

`make snapshot` produces the same artifacts under `./dist/` without
publishing — useful for verifying the build before tagging.

## Continuous integration

`.github/workflows/ci.yml` runs `make test` and `make lint` on every PR
against `main` and every push to `main`. Workflow YAML is linted by
`actionlint`.
