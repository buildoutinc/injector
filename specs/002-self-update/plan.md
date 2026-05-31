# Implementation Plan: Self-Update via GitHub Releases

**Branch**: `002-self-update` | **Date**: 2026-05-30 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-self-update/spec.md`

## Summary

Three coordinated pieces deliver the spec:

1. **`inject upgrade` subcommand** — wired with
   `github.com/rhysd/go-github-selfupdate/selfupdate`. Compares the
   running binary's semver to the latest GitHub Release for
   `buildoutinc/injector`, downloads the matching archive, verifies its
   checksum, and atomically swaps the binary on disk. Supports
   `--check` (no-op preview), `--pre-release` (opt-in to pre-releases),
   and refuses to write when the binary path isn't user-writable.
2. **Passive update check** — `inject version`, root help (`inject`,
   `inject --help`, `inject help`) print a one-line "newer version
   available" notice when applicable. Backed by a 24h cache file under
   the user's cache directory, executed concurrently with the
   foreground command so it never blocks output past a 200ms budget;
   disabled by `INJECT_NO_UPDATE_CHECK=1`.
3. **Release pipeline** — A separate `release.yml` GitHub Actions
   workflow triggers on `v*.*.*` tags pushed to the default branch.
   GoReleaser builds `linux/{amd64,arm64}` + `darwin/{amd64,arm64}`
   archives whose names match the self-update library's expected
   pattern (`inject_<version>_<os>_<arch>.tar.gz`), embeds the version
   into the binary via `-ldflags -X main.version=`, generates release
   notes grouped by Conventional Commit type with `git-chglog`, and
   uploads everything (binaries + `checksums.txt`) to the GitHub
   Release.

The existing `.goreleaser.yaml` from feature 001 already produces
asset names of the form `inject_<version>_<os>_<arch>.tar.gz`; this
matches go-github-selfupdate's `(_|-){os}(_|-){arch}` regex matcher,
so no naming rework is required — only release-trigger and changelog
configuration changes.

## Technical Context

**Language/Version**: Go 1.26.3 (unchanged from feature 001).

**Primary Dependencies**:
- `github.com/rhysd/go-github-selfupdate/selfupdate` — self-update engine
- `github.com/Masterminds/semver/v3` (transitive via selfupdate) for semver compare
- Existing: `github.com/alecthomas/kong`, `github.com/charmbracelet/bubbletea`, `github.com/carapace-sh/carapace`

**Storage**:
- Update-check cache file at `${XDG_CACHE_HOME:-~/.cache}/inject/update-check.json` on Linux, `~/Library/Caches/inject/update-check.json` on macOS (use `os.UserCacheDir()`). One small JSON file: `{checked_at, latest_version}`.

**Testing**:
- Unit tests with the selfupdate library hidden behind an in-process `Updater` interface stubbed by a fake — no network in `go test ./...` (Constitution VI).
- Workflow validation: `actionlint` in CI; release pipeline itself only verifiable by pushing a tag (covered in quickstart).

**Target Platform**: macOS + Linux on `amd64` + `arm64`. Windows out of scope (consistent with feature 001).

**Project Type**: CLI binary; same module as before (`github.com/buildoutinc/injector`).

**Performance Goals**:
- Passive update check adds ≤ 200ms wall-clock to `inject version`, `inject`, `inject help` even when the network is unreachable (SC-004). Implementation enforces this via a 200ms `context.WithTimeout` on the check goroutine; if the foreground command finishes first, the notice is silently dropped.
- `inject upgrade` end-to-end completes in < 30s on broadband (SC-001).

**Constraints**:
- Constitution VI: tests run offline — selfupdate must be invoked through a swappable Updater so tests use a fake.
- No mandatory GitHub auth for end users (assumption from spec); reads `GITHUB_TOKEN` from env if present to lift rate limits.
- Permission pre-check via `unix.Access(path, W_OK)` (Linux/macOS only — fine).
- Notice must be on stdout (FR-013) but visually distinct so machine readers ignore it: prefix with a blank line and `==>` marker, à la Homebrew.

**Scale/Scope**: Single user per machine; one cache file; ≤ 1 GitHub API request per 24h per user (SC-005).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Pluggable Backend Architecture | ✅ N/A | Self-update is CLI plumbing; touches no backend interface. |
| II. Inheritance Hierarchy & Multi-Tenancy | ✅ N/A | No org/project/service/env modeling here. |
| III. Git-Centric Persistence & Schema Validation | ✅ N/A | The update-check cache is per-user state, not project data — not subject to YAML/schema/Git rules. |
| IV. Security & Least Privilege | ✅ PASS | Checksums verified before binary swap (FR-004); permission pre-check (FR-006) prevents writing where the user lacks rights; no privileged operations; no secrets handled; opt-out env var available (FR-012). |
| V. Developer Experience & CLI Design | ✅ PASS | `inject upgrade --help`, `inject version` follow the git-like convention; the notice is a one-line nudge consistent with "convention over configuration" (Principle V). |
| VI. Testing & Quality — Local Isolation | ✅ PASS | All tests use a stubbed `Updater` and a temp-dir cache — no network. Release workflow itself is exercised by pushing a tag, which is explicit opt-in (analogous to the build-tag-gated e2e suite). |

**Result**: All gates pass.

## Project Structure

### Documentation (this feature)

```text
specs/002-self-update/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── cli-commands.md         # inject upgrade, inject version, notice contract
│   └── release-artifacts.md    # asset naming + checksums file format
└── tasks.md                    # generated later by /speckit-tasks
```

### Source Code (repository root)

```text
internal/cli/
├── upgrade.go              # UpgradeCmd: drives the Updater
├── upgrade_test.go         # tests using FakeUpdater
├── version.go              # VersionCmd; prints version then renders notice (if any)
├── version_test.go
├── help.go                 # HelpCmd: explicit `inject help` alias for `--help`
├── notice.go               # update-check goroutine + 200ms budget + cache I/O
├── notice_test.go          # asserts: never blocks; cache hit avoids network; opt-out honored
└── (existing: root.go, project.go, project_init.go, signal.go, completion.go, tui.go, execute.go)

internal/updater/
├── updater.go              # Updater interface + selfupdate-backed implementation
├── fake.go                 # FakeUpdater (in-memory) for tests
└── updater_test.go

internal/updatecheck/
├── cache.go                # ReadCache / WriteCache to user cache dir
├── cache_test.go
├── check.go                # async Check(ctx) with timeout; returns Notice (or nil)
└── check_test.go

cmd/inject/main.go          # unchanged except: pass `version` ldflag value
.github/workflows/
├── ci.yml                  # unchanged from feature 001
└── release.yml             # NEW — tagged release pipeline (GoReleaser + chglog)
.chglog/
├── config.yml              # git-chglog grouping config (Conventional Commits)
└── CHANGELOG.tpl.md
.goreleaser.yaml            # updated: ldflags inject version, archives confirm naming
```

**Structure Decision**: Three small internal packages — `internal/updater`,
`internal/updatecheck`, plus extensions to existing `internal/cli` —
keep the new plumbing isolated and testable through interfaces.
`internal/cli` owns user-facing wiring (subcommands, notice rendering);
`internal/updater` owns the selfupdate library boundary so tests can
substitute a fake; `internal/updatecheck` owns the 24h cache + async
check goroutine. The release pipeline lives entirely under
`.github/workflows/release.yml` + `.chglog/` + `.goreleaser.yaml`.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Wrapping `selfupdate` in an `Updater` interface (`internal/updater`) | Constitution VI requires offline tests; selfupdate's package-level helpers hit api.github.com directly. The interface lets `upgrade_test.go` and `notice_test.go` use a fake. | Calling `selfupdate.UpdateSelf` directly would force a real network call in unit tests, violating Principle VI. |
| Async update check goroutine with 200ms timeout | SC-004 caps the latency cost of the notice at 200ms even on a network failure. A synchronous check would either block longer than that or skip entirely on a slow network. | Synchronous check with a tight timeout would still block the foreground command's stdout flush; the async approach hides the latency by racing the foreground command. |
| Adding `git-chglog` as a CI-time tool (alongside GoReleaser) | GoReleaser's built-in changelog groups commits by Conventional Commit type but is template-locked to GoReleaser's default sections; the team wants explicit Features / Fixes / Other sections and tolerant handling of un-prefixed commits (FR-017). `git-chglog` gives us a small config file to express that exactly. | GoReleaser-only changelog works but doesn't surface un-prefixed commits in a named section by default — they'd disappear into "Other Changes" with no header customization. |
