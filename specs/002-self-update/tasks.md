---
description: "Task list for 002-self-update"
---

# Tasks: Self-Update via GitHub Releases

**Input**: Design documents from `/specs/002-self-update/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Tests are first-class — Constitution VI requires offline coverage of every behavior the feature introduces. Test tasks below use a fake `Updater` and a temp-dir cache; no network is touched by `go test ./...`.

**Organization**: Tasks are grouped by user story (US1 = `inject upgrade`, US2 = passive notice, US3 = release pipeline). Each phase is independently testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Different files, no dependency on incomplete tasks → parallelizable
- **[Story]**: Maps task to its user story (US1, US2, US3)
- Paths are relative to `/Users/shane/src/injector`

---

## Phase 1: Setup (Shared Infrastructure)

- [X] T001 Add the self-update dependency: `go get github.com/rhysd/go-github-selfupdate/selfupdate`; commit resulting `go.mod` / `go.sum`
- [X] T002 [P] Create the new package directories: `internal/updater/` and `internal/updatecheck/`
- [X] T003 [P] Update `.goreleaser.yaml`'s `builds[0].ldflags` to include `-X main.version=v{{ .Version }}` (so the `v` prefix is baked in) and confirm `archives[0].name_template` remains `{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}` (matches the self-update library's `_<os>_<arch>` regex per research §2)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared abstractions that both US1 and US2 depend on.

**⚠️ CRITICAL**: US1 and US2 cannot start until this phase is complete. US3 can run independently in parallel.

- [X] T004 In `internal/updater/updater.go`, declare the `Updater` interface (`Latest(ctx, LatestOpts) (Release, error)` and `Apply(ctx, Release, binPath) (Result, error)`) plus the `LatestOpts`, `Release`, `Result` value types from `data-model.md`
- [X] T005 In `internal/updater/updater.go`, implement `NewGithub(slug, currentVersion string) Updater` that constructs a `selfupdate.Updater` via `selfupdate.NewUpdater(selfupdate.Config{Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"}})`, mapping `LatestOpts.IncludePrereleases` to the config's `Prerelease` field; `Latest` calls `DetectLatest(slug)`, `Apply` calls `UpdateTo(release, binPath)`
- [X] T006 [P] In `internal/updater/fake.go`, implement `FakeUpdater` (in-memory) for tests: configurable `LatestFunc`, `ApplyFunc`, and a written-bytes recorder so tests can verify the swap path was taken
- [X] T007 [P] In `internal/updater/updater_test.go`, add `TestFakeUpdater_RoundTrip` that exercises the fake to lock the test contract (separate from the real GitHub-backed implementation)
- [X] T008 In `cmd/inject/main.go`, ensure the `version` package-level var is wired into `cli.Execute` (already in place from feature 001) and add a `Commit` / `Date` var pair populated via the same ldflag mechanism (defaults `"none"` / `"unknown"`)
- [X] T009 [P] In `internal/cli/version.go`, define `func Version() string` returning the linker-provided version and `func BuildInfo() (commit, date string)` so other packages can format the version line without importing `main`

**Checkpoint**: `go build ./...` passes; `go test ./internal/updater/...` is green; nothing user-visible yet.

---

## Phase 3: User Story 1 — One-Command Self-Upgrade (Priority: P1) 🎯 MVP

**Goal**: `inject upgrade` (with `--check` and `--pre-release` variants) installs the latest release or refuses cleanly per the cli-commands contract.

**Independent Test**: `./bin/inject upgrade --check` returns 0 against a `FakeUpdater` that advertises a newer version, then `./bin/inject upgrade` reports a successful swap; both flows hit no network.

### Tests for User Story 1

- [X] T010 [P] [US1] In `internal/cli/upgrade_test.go`, add `TestUpgrade_AlreadyCurrent` — wire `UpgradeCmd` with a `FakeUpdater` whose `Latest` returns the same version; assert stdout `"inject: already up to date (vX.Y.Z)\n"`, no disk writes, exit 0
- [X] T011 [P] [US1] In `internal/cli/upgrade_test.go`, add `TestUpgrade_NewerVersion` — fake returns a strictly newer release; assert stdout `"inject: upgraded vA.B.C → vX.Y.Z\n"`, `Apply` was called once, exit 0
- [X] T012 [P] [US1] In `internal/cli/upgrade_test.go`, add `TestUpgrade_NetworkError` — fake `Latest` returns an error; assert stderr starts with `"inject: cannot upgrade: "`, no `Apply` call, exit non-zero
- [X] T013 [P] [US1] In `internal/cli/upgrade_test.go`, add `TestUpgrade_ChecksumMismatch` — fake `Apply` returns a checksum-mismatch sentinel; assert stderr contains "checksum mismatch", binary untouched on disk (fake records 0 bytes written), exit non-zero
- [X] T014 [P] [US1] In `internal/cli/upgrade_test.go`, add `TestUpgrade_PermissionDenied` — point `UpgradeCmd.binPath` at a tempdir file with mode `0444` on a read-only directory; assert stderr names the path and suggests a package manager, no `Latest` call (pre-check happens first), exit non-zero
- [X] T015 [P] [US1] In `internal/cli/upgrade_test.go`, add `TestUpgrade_CheckOnly` — fake advertises newer version; run with `--check`; assert stdout `"A newer version is available: vX.Y.Z (you have vA.B.C)\n"`, fake `Apply` not called, exit 0
- [X] T016 [P] [US1] In `internal/cli/upgrade_test.go`, add `TestUpgrade_PreReleaseFlag` — assert that `--pre-release` causes `Updater.Latest` to be called with `IncludePrereleases=true`
- [X] T017 [P] [US1] In `internal/cli/upgrade_test.go`, add `TestUpgrade_UnsupportedPlatform` — fake returns a "no asset for GOOS/GOARCH" error; assert stderr names the platform, exit non-zero (SC-003)

### Implementation for User Story 1

- [X] T018 [US1] In `internal/updater/permission.go`, implement `func IsWritable(path string) error` using `golang.org/x/sys/unix.Access(path, unix.W_OK)`; on EACCES/EROFS return a typed error that carries the path so the CLI can format the remediation message
- [X] T019 [US1] In `internal/cli/upgrade.go`, define `type UpgradeCmd struct { Check bool `+"`"+`help:"Only report whether an upgrade is available; do not modify anything."`+"`"+`; PreRelease bool `+"`"+`name:"pre-release" help:"Include pre-release tags."`+"`"+` }` plus `Run(ctx context.Context, u updater.Updater, out io.Writer, stderr io.Writer) error` orchestrating: resolve `os.Executable()` → permission pre-check → `u.Latest()` → semver compare → `--check` branch vs apply → render outcomes per the cli-commands contract
- [X] T020 [US1] In `internal/cli/root.go`, add `Upgrade UpgradeCmd `+"`"+`cmd:"" help:"Upgrade the inject binary to the latest GitHub release."`+"`"+`` to `RootCmd`
- [X] T021 [US1] In `internal/cli/execute.go`, bind a default `updater.Updater` (from `updater.NewGithub("buildoutinc/injector", cli.Version())`) via `kong.BindTo(...)` so tests can override it; bind `io.Writer` already exists from feature 001

**Checkpoint**: `./bin/inject upgrade --check` and `./bin/inject upgrade` work end-to-end against the fake (unit tests); against the real GitHub if a release exists.

---

## Phase 4: User Story 2 — Passive "Newer Version Available" Notice (Priority: P1)

**Goal**: `inject`, `inject help`, `inject --help`, and `inject version` print a one-line notice when a newer release is available; the check is rate-limited, async (200ms budget), opt-out-able, and never fails the foreground command.

**Independent Test**: With `INJECT_NO_UPDATE_CHECK` unset and a 0-day-old cache that says `latest_version` is newer, `./bin/inject version` prints the version line followed by the `==>` notice block; with `INJECT_NO_UPDATE_CHECK=1` set, no notice; with the cache absent and a `FakeUpdater` returning an error, no notice and exit 0.

### Tests for User Story 2

- [X] T022 [P] [US2] In `internal/updatecheck/cache_test.go`, add `TestCache_RoundTrip` — write a cache record to a tempdir, read it back, assert equality including RFC3339 timestamp parsing
- [X] T023 [P] [US2] In `internal/updatecheck/cache_test.go`, add `TestCache_MalformedReturnsCacheMiss` — write garbage to the cache file; `ReadCache` returns `(zero, nil)` (treated as miss, not error)
- [X] T024 [P] [US2] In `internal/updatecheck/check_test.go`, add `TestCheck_RespectsOptOut` — set `INJECT_NO_UPDATE_CHECK=1`; assert `Check` returns `(nil, nil)` immediately, no `Updater.Latest` call
- [X] T025 [P] [US2] In `internal/updatecheck/check_test.go`, add `TestCheck_FreshCacheSkipsNetwork` — write a cache record 1h old; assert `Updater.Latest` is **not** called; assert the returned `Notice` reflects the cached `latest_version` when newer
- [X] T026 [P] [US2] In `internal/updatecheck/check_test.go`, add `TestCheck_StaleCacheTriggersNetwork` — write a cache record 25h old; fake returns a newer version; assert `Updater.Latest` called once; cache file is updated; `Notice` reflects the new latest
- [X] T027 [P] [US2] In `internal/updatecheck/check_test.go`, add `TestCheck_NetworkErrorIsSwallowed` — stale cache + fake returning error; assert `Check` returns `(nil, nil)` (no error to caller); existing cache file is left untouched
- [X] T028 [P] [US2] In `internal/updatecheck/check_test.go`, add `TestCheck_DevVersionSkipsCheck` — pass `binaryVersion = "dev"`; assert no `Updater.Latest` call, returns nil notice
- [X] T029 [P] [US2] In `internal/updatecheck/check_test.go`, add `TestCheck_Timeout` — fake `Updater.Latest` blocks on a channel; call `Check` with a 50ms ctx; assert the call returns within 200ms total with nil notice and no panic
- [X] T030 [P] [US2] In `internal/cli/notice_test.go`, add `TestRenderNotice_TTY` and `TestRenderNotice_NoTTY` — write to a `bytes.Buffer`; assert format `"\n==> A newer version is available: vX.Y.Z\n==> Run \"inject upgrade\" to install it.\n"` (no ANSI escapes in the non-TTY case)
- [X] T031 [P] [US2] In `internal/cli/version_test.go`, add `TestVersionCommand_NoticeRendered` — install a fake `updatecheck.Checker` that returns a notice; assert stdout contains the version line then the notice block; exit 0
- [X] T032 [P] [US2] In `internal/cli/version_test.go`, add `TestVersionCommand_NoNoticeWhenUpToDate` — fake returns nil; stdout is only the version line; exit 0
- [X] T033 [P] [US2] In `internal/cli/version_test.go`, add `TestVersionCommand_NoticeNeverBlocksLongerThan200ms` — fake checker blocks 5s; `VersionCmd.Run` must return within 250ms with no notice
- [X] T034 [P] [US2] In `internal/cli/notice_test.go`, add `TestNoticeAppendedToHelp` — drive root help via `cli.Execute(...,"--help")` with a notice-emitting checker; assert the notice block follows the help screen on stdout

### Implementation for User Story 2

- [X] T035 [US2] In `internal/updatecheck/cache.go`, implement `type Record struct { CheckedAt time.Time `+"`"+`json:"checked_at"`+"`"+`; LatestVersion string `+"`"+`json:"latest_version"`+"`"+`; BinaryVersion string `+"`"+`json:"binary_version"`+"`"+` }`, `ReadCache(dir string) (Record, bool, error)` (bool=present), and `WriteCache(dir string, r Record) error` using `os.WriteFile` to a `tmp+rename` (atomic on POSIX)
- [X] T036 [US2] In `internal/updatecheck/cache.go`, implement `func DefaultCacheDir() (string, error)` returning `filepath.Join(os.UserCacheDir(), "inject")` and ensuring it exists with `0700`
- [X] T037 [US2] In `internal/updatecheck/check.go`, define `type Checker interface { Start(ctx context.Context); Notice() *Notice }` and a concrete `BackgroundChecker` that:
  - on `Start`, returns immediately if `INJECT_NO_UPDATE_CHECK` is set or `binaryVersion == "dev"`,
  - reads the cache; if fresh (< 24h), populates the notice channel from the cached `latest_version` and returns,
  - otherwise spawns a goroutine that calls `Updater.Latest(ctx with 200ms timeout)`, writes the result to a 1-buffered chan, and updates the cache file on success.
  `Notice()` non-blockingly drains the channel (returns nil if the deadline has passed or the check is still running).
- [X] T038 [US2] In `internal/cli/notice.go`, implement `func RenderNotice(w io.Writer, n *updatecheck.Notice, isTTY bool)` producing the exact two-line `==>` block from cli-commands.md; isTTY adds ANSI cyan
- [X] T039 [US2] In `internal/cli/version.go`, define `type VersionCmd struct{}` and `Run(ctx context.Context, c updatecheck.Checker, out io.Writer) error` that (a) calls `c.Start(ctx)` first, (b) prints `inject <version> (commit <sha>, built <date>)`, (c) calls `RenderNotice(out, c.Notice(), isTTYOf(out))`, (d) exits 0
- [X] T040 [US2] In `internal/cli/root.go`, add `Version VersionCmd `+"`"+`cmd:"" help:"Show inject version."`+"`"+`` to `RootCmd`; remove or merge the existing `kong.VersionFlag` so `inject --version` and `inject version` are both routed through `VersionCmd` (single source of truth)
- [X] T041 [US2] In `internal/cli/help.go`, implement `type HelpCmd struct{}` mapped to `inject help` that simply re-runs the parser with `--help` so the same help screen + notice surface
- [X] T042 [US2] In `internal/cli/execute.go`, instantiate the default `updatecheck.BackgroundChecker` (using the same `updater.Updater` from US1) and `kong.BindTo` it as `updatecheck.Checker`; in the help-rendering paths (no-args case and `kong.UsageOnError` retry), call `c.Start(ctx)` once before kong prints, and call `RenderNotice` once after the help body is written
- [X] T043 [US2] In `internal/cli/execute.go`, ensure the notice is rendered **at most once per invocation** by guarding with a `sync.Once` shared across the help path and the subcommand path

**Checkpoint**: All US2 tests green; running `./bin/inject version` (with a temp cache + fake checker) prints the version line and the notice block.

---

## Phase 5: User Story 3 — Automated Tagged GitHub Release (Priority: P1)

**Goal**: Pushing a `v*.*.*` tag triggers a release.yml workflow that uses `git-chglog` for grouped notes and GoReleaser for the four-archive matrix + `checksums.txt`, all attached to the GitHub Release for that tag.

**Independent Test**: Locally run `goreleaser release --snapshot --clean` and confirm `dist/` contains the four `inject_<version>_<os>_<arch>.tar.gz` files plus `checksums.txt`; run `git-chglog --config .chglog/config.yml --next-tag v0.X.0 v0.X.0` and inspect the rendered Markdown.

### Implementation for User Story 3

- [X] T044 [P] [US3] Create `.chglog/config.yml` with `style: github`, `template: CHANGELOG.tpl.md`, and an `options.commit_groups.group_by: Type` block declaring titles for `feat → Features`, `fix → Fixes`, `perf → Performance`, `refactor → Refactor`, `docs → Documentation`; set `options.commits.filters.Type` to those types so they are included; leave unmatched commits to fall through to the template's catch-all
- [X] T045 [P] [US3] Create `.chglog/CHANGELOG.tpl.md` rendering, for the current tag only, sections per `CommitGroups` in the declared order, followed by an `## Other Changes` section iterating `Commits` not in any group (covers FR-017)
- [X] T046 [P] [US3] In `.goreleaser.yaml`, set `release.header: ""` and `release.body: ""` to defer to the workflow-provided body (loaded from `git-chglog` output via `--release-notes` CLI flag); add `release.prerelease: auto` so `vX.Y.Z-rc.N` tags become pre-releases
- [X] T047 [US3] Create `.github/workflows/release.yml`:
  - Triggers: `on: { push: { tags: ['v*.*.*'] } }`; top-level `permissions: { contents: write }`
  - Single job `release` with `runs-on: ubuntu-latest` and `if: github.repository == 'buildoutinc/injector'`
  - Steps: `actions/checkout@v4` with `fetch-depth: 0` (chglog needs full history); `actions/setup-go@v5` with `go-version-file: go.mod`; install `git-chglog` (`go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest`); validate tag shape with a 6-line `bash -e` step that `grep -E`s the regex from contracts/release-artifacts.md and fails fast on mismatch (FR-020); generate notes (`git-chglog -o /tmp/notes.md $GITHUB_REF_NAME`); run `goreleaser release --clean --release-notes /tmp/notes.md` with `env: { GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }} }`
- [X] T048 [P] [US3] In the README's release section (created in T053 below), reference the manual command equivalent: `git-chglog -o NOTES.md vX.Y.Z && goreleaser release --clean --release-notes NOTES.md` for maintainers cutting a release locally

**Checkpoint**: `goreleaser release --snapshot --clean` produces the expected four archives + checksums; `git-chglog` renders the grouped Markdown; `release.yml` passes `actionlint`.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T049 [P] In `Makefile`, add a `snapshot` target with `## Build a local snapshot release (no publish).` invoking `goreleaser release --snapshot --clean`; add a `notes` target `## Render the changelog for the latest tag.` invoking `git-chglog -o /tmp/notes.md $(git describe --tags --abbrev=0)`; both target additions follow the existing `## help` convention
- [X] T050 In `.github/workflows/ci.yml`, append an `actionlint` step (`reviewdog/action-actionlint@v1` or equivalent) so future workflow edits are validated on every PR
- [X] T051 In `cmd/inject/main_test.go`, extend `TestBinarySmoke_HelpAndInit` with two new subtests: `version subcommand exits 0 and prints "inject "` and `upgrade --check exits 0 (offline-safe with INJECT_NO_UPDATE_CHECK=1)` so the smoke test exercises the new user-visible surface end-to-end
- [X] T052 [P] Run `make tidy`, then `make test` and `make lint` from a clean checkout; confirm zero issues and `go.sum` minimal
- [X] T053 Update `README.md` with three new sections (per FR-022): **Upgrading** (`inject upgrade`, `inject upgrade --check`, `INJECT_NO_UPDATE_CHECK`, where the cache lives), **Releasing** (Conventional Commit conventions, tag-and-push flow, what release.yml does), and **Release artifacts** (the four archive names, what's inside each, the `checksums.txt` contents)
- [X] T054 Walk the quickstart end-to-end (Sections A, D, F at minimum — Sections B/C/E require a real release and are validated only when actually cutting one); check off each expected outcome

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: no dependencies
- **Foundational (Phase 2)**: depends on Phase 1; blocks US1 and US2 (NOT US3)
- **US1 (Phase 3)**: depends on Phase 2
- **US2 (Phase 4)**: depends on Phase 2; can run in parallel with US1
- **US3 (Phase 5)**: depends only on Phase 1 (T003 specifically); independent of US1/US2 implementation — can be done by a different contributor at the same time
- **Polish (Phase 6)**: depends on US1, US2, US3

### Within each user story

- US1 tests (T010–T017) and the supporting fake (T006) can be written before T019/T020 land; they will fail until then.
- US2 cache & checker tests (T022–T029) only depend on T035–T037; rendering tests (T030–T034) only depend on T038–T042.

### Parallel opportunities

- Setup: T002 and T003 in parallel
- Foundational: T006/T007 and T009 in parallel after T004/T005
- US1: all of T010–T017 in parallel (different test functions, one file each)
- US2: all of T022–T034 in parallel (different test files)
- US3: T044/T045/T046 in parallel (different files); T047 depends on T044+T045
- Different contributors can own US1, US2, US3 in parallel after Phase 2 starts.

---

## Parallel Example: User Story 1 tests

```bash
# All US1 test functions live in a single file but are independent test funcs.
# Author them concurrently (different functions don't conflict):
Task: "TestUpgrade_AlreadyCurrent in internal/cli/upgrade_test.go"
Task: "TestUpgrade_NewerVersion in internal/cli/upgrade_test.go"
Task: "TestUpgrade_NetworkError in internal/cli/upgrade_test.go"
Task: "TestUpgrade_ChecksumMismatch in internal/cli/upgrade_test.go"
Task: "TestUpgrade_PermissionDenied in internal/cli/upgrade_test.go"
Task: "TestUpgrade_CheckOnly in internal/cli/upgrade_test.go"
Task: "TestUpgrade_PreReleaseFlag in internal/cli/upgrade_test.go"
Task: "TestUpgrade_UnsupportedPlatform in internal/cli/upgrade_test.go"
```

---

## Implementation Strategy

### MVP scope (recommended)

1. Phase 1 (Setup) — 3 tasks
2. Phase 2 (Foundational) — 6 tasks
3. Phase 3 (US1) — `inject upgrade` works
4. Phase 4 (US2) — passive notice works
5. Phase 5 (US3) — release.yml works

All three user stories are P1 — there is no smaller MVP. If you must
ship in increments:

1. Setup + Foundational → no user-visible change yet
2. US1 alone → `inject upgrade` works but no nudge; users must discover it
3. US2 → users get the nudge; the upgrade command they're nudged to use already exists
4. US3 → without it, US1 and US2 work against the v0.1.0 release but never see anything newer; cutting a release closes the loop
5. Polish → README + Makefile niceties + CI workflow lint

### Parallel team strategy

After Phase 2:
- Developer A: US1 (Phase 3)
- Developer B: US2 (Phase 4)
- Developer C: US3 (Phase 5) — does not need Phase 2 at all; can start as soon as T003 lands

---

## Notes

- [P] = different files, no dependency on incomplete tasks
- Tests use `FakeUpdater` + tempdir cache → fully offline (Constitution VI)
- `INJECT_NO_UPDATE_CHECK=1` is the universal escape hatch and is honored on every command, including `version` and the help paths
- `inject upgrade` itself can hit the network when explicitly invoked — it's the user's action, not a passive check
- The release pipeline (US3) is YAML-only and can be reviewed/merged before the Go code lands if convenient
