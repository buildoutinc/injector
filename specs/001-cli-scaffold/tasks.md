---
description: "Task list for 001-cli-scaffold"
---

# Tasks: CLI Scaffold & Build Pipeline

**Input**: Design documents from `/specs/001-cli-scaffold/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-commands.md, quickstart.md

**Tests**: Tests are first-class deliverables for this feature — spec User
Story 4 and FR-020 make them required, not optional.

**Organization**: Tasks are grouped by user story so each story can be
implemented and validated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Different files, no dependency on incomplete tasks → parallelizable
- **[Story]**: Maps task to its user story (US1, US2, US3, US4, US5)
- All paths are repo-relative from `/Users/shane/src/injector`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize the Go module's source layout and pin dependencies.

- [X] T001 Create directory skeleton: `cmd/inject/`, `internal/cli/`, `.github/workflows/` at repo root
- [X] T002 Add CLI dependencies to `go.mod`: run `go get github.com/alecthomas/kong github.com/charmbracelet/bubbletea github.com/carapace-sh/carapace` and commit the resulting `go.mod`/`go.sum`
- [X] T003 [P] Add `.golangci.yaml` at repo root with a minimal preset (`govet`, `staticcheck`, `errcheck`, `ineffassign`, `unused`)
- [X] T004 [P] Add `.gitignore` entries for `/bin/`, `/dist/`, and `coverage.out` (append; do not clobber existing `.gitignore`)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Wiring shared by every subcommand — the cancelable context, the
kong application bootstrap, and the bubbletea runtime hook. Every story
below depends on these.

**⚠️ CRITICAL**: No user-story phase can start until this phase is complete.

- [X] T005 Implement signal-aware context helper in `internal/cli/signal.go` exporting `func NotifyContext(parent context.Context) (context.Context, context.CancelFunc)` wrapping `signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)`
- [X] T006 Implement root command grammar in `internal/cli/root.go`: `type RootCmd struct { Project ProjectCmd \`cmd:"" help:"Manage Injector projects."\`; Version kong.VersionFlag \`help:"Show version."\` }` plus a `Description` constant for the kong help
- [X] T007 Implement program entrypoint in `cmd/inject/main.go`: build root context via `internal/cli.NotifyContext`, call `kong.Parse(&cli.RootCmd{}, kong.Name("inject"), kong.Description(cli.Description), kong.UsageOnError(), kong.Bind(ctx))`, invoke `ctx.Run(...)`, map `context.Canceled` → exit code 130 and other errors → exit code 1
- [X] T008 [P] Wire bubbletea import hook in `internal/cli/tui.go` exporting `func Run(ctx context.Context, m tea.Model) error` that calls `tea.NewProgram(m, tea.WithContext(ctx)).Run()` (unused this slice; reserved for future subcommands)
- [X] T009 [P] Register carapace completion in `internal/cli/completion.go` exposing a hidden `_carapace` subcommand on `RootCmd` that emits per-shell completion scripts

**Checkpoint**: `go build ./...` succeeds; `./bin/inject --help` would render kong's default help (no real subcommands yet).

---

## Phase 3: User Story 1 - Discoverable CLI Entry Point (Priority: P1) 🎯 MVP

**Goal**: Running `inject`, `inject --help`, or `inject -h` prints a help
screen listing every subcommand. Unknown subcommands print to stderr and
exit non-zero.

**Independent Test**: Build `./bin/inject`; run `./bin/inject`,
`./bin/inject --help`, `./bin/inject -h`, and `./bin/inject does-not-exist`.
First three exit 0 with help on stdout listing `project`; the last exits
non-zero with an error on stderr.

### Tests for User Story 1

- [X] T010 [P] [US1] In `internal/cli/root_test.go`, add `TestRootHelp_NoArgs` and `TestRootHelp_HelpFlag` — invoke kong in-process with `kong.Writers(stdout, stderr)` and `[]string{}` / `[]string{"--help"}`; assert exit 0 path, stdout contains `inject`, `project`, `Commands`, `Flags`
- [X] T011 [P] [US1] In `internal/cli/root_test.go`, add `TestRootUnknownSubcommand` — pass `[]string{"does-not-exist"}`; assert kong returns a non-nil error, stderr is non-empty, stdout is empty

### Implementation for User Story 1

- [X] T012 [US1] In `internal/cli/root.go`, set the kong `Description` to a one-line tool summary and ensure `Project` subcommand has a `help:` tag so it shows up in `Commands:` (verified by T010)
- [X] T013 [US1] In `cmd/inject/main.go`, confirm `kong.UsageOnError()` is wired so unknown-subcommand errors print usage to stderr (verified by T011)

**Checkpoint**: US1 tests green; `./bin/inject` and `./bin/inject --help` render help.

---

## Phase 4: User Story 2 - Stubbed `project init` Subcommand (Priority: P1)

**Goal**: `inject project init` prints exactly `TODO: init!\n` to stdout
and exits 0. `inject project` and `inject project --help` print
project-scoped help listing `init`.

**Independent Test**: Build `./bin/inject`; run `./bin/inject project init`
and confirm stdout is exactly `TODO: init!\n`, stderr empty, exit 0. Run
`./bin/inject project` and confirm help mentions `init`.

### Tests for User Story 2

- [X] T014 [P] [US2] In `internal/cli/project_test.go`, add `TestProjectHelp` and `TestProjectHelpFlag` — invoke kong with `[]string{"project"}` and `[]string{"project", "--help"}`; assert exit 0, stdout contains `init`
- [X] T015 [P] [US2] In `internal/cli/project_init_test.go`, add `TestProjectInit_Output` — invoke kong with `[]string{"project", "init"}` binding a test `context.Background()`; assert stdout equals `"TODO: init!\n"`, stderr empty, error nil
- [X] T016 [P] [US2] In `internal/cli/project_init_test.go`, add `TestProjectInitHelpFlag` — invoke kong with `[]string{"project", "init", "--help"}`; assert exit 0, stdout contains `init`

### Implementation for User Story 2

- [X] T017 [US2] In `internal/cli/project.go`, define `type ProjectCmd struct { Init ProjectInitCmd \`cmd:"" help:"Scaffold a new Injector project."\` }` with `help:` text suitable for the `Commands:` listing
- [X] T018 [US2] In `internal/cli/project_init.go`, define `type ProjectInitCmd struct{}` and implement `func (c *ProjectInitCmd) Run(ctx context.Context, w io.Writer) error { _, err := io.WriteString(w, "TODO: init!\n"); return err }` — accept stdout writer via `kong.Bind(os.Stdout)` so tests can substitute a buffer
- [X] T019 [US2] In `cmd/inject/main.go`, add `kong.Bind(os.Stdout)` so `ProjectInitCmd.Run` receives the real stdout in production

**Checkpoint**: US1 + US2 tests green; `./bin/inject project init` prints `TODO: init!`.

---

## Phase 5: User Story 3 - Self-Documenting Build & Release Makefile (Priority: P1)

**Goal**: A self-documenting `Makefile` at the repo root with
`help`/`build`/`test`/`lint`/`release`/`tidy`/`clean` targets. `make` or
`make help` lists every target with its inline `##` description.

**Independent Test**: From a fresh clone, run `make` (expect help listing
all targets), `make build` (expect `./bin/inject` produced), `make test`
(passes), `make lint` (passes).

### Implementation for User Story 3

- [X] T020 [US3] Create `Makefile` at repo root with `.PHONY` declarations and an awk-based `help` target that parses inline `##` comments (see research.md §5). Default goal: `help`
- [X] T021 [US3] Add `build` target to `Makefile`: `go build -o ./bin/inject ./cmd/inject` with `## Build ./bin/inject for the host platform.`
- [X] T022 [US3] Add `test` target to `Makefile`: `go test ./...` with `## Run the full Go test suite locally.`
- [X] T023 [US3] Add `lint` target to `Makefile`: `golangci-lint run ./...` with `## Run golangci-lint against the codebase.`
- [X] T024 [US3] Add `tidy` and `clean` targets to `Makefile` with `## Run go mod tidy.` and `## Remove build artifacts.` (deletes `./bin` and `./dist`)
- [X] T025 [US3] Create `.goreleaser.yaml` at repo root configured for `darwin/{amd64,arm64}` and `linux/{amd64,arm64}` archives, binary name `inject`, main `./cmd/inject`, checksums on, GitHub release enabled
- [X] T026 [US3] Add `release` target to `Makefile`: `goreleaser release --clean` with `## Build cross-platform archives and publish to GitHub Releases.`

**Checkpoint**: `make help` lists every target; `make build`, `make test`, `make lint` all exit 0 on the current tree.

---

## Phase 6: User Story 4 - Automated Test Suite Runnable from Make (Priority: P1)

**Goal**: The repo ships a `go test` suite that runs entirely offline and
exits non-zero on regressions in the CLI surface.

**Independent Test**: Run `make test` on a clean checkout (passes). Edit
`project_init.go` to print `oops` instead of `TODO: init!`, re-run `make
test` — at least one test fails and exit is non-zero. Revert.

> Note: Most coverage was already authored in US1/US2. This phase adds the
> end-to-end smoke test that exercises the wired-up binary and confirms
> the suite as a whole satisfies FR-020.

### Tests for User Story 4

- [X] T027 [P] [US4] In `cmd/inject/main_test.go`, add `TestBinarySmoke_HelpAndInit` guarded by `if testing.Short() { t.Skip() }` — build the binary into a temp dir via `go build`, then exec it with `[]string{}` (assert exit 0 + stdout contains `project`), `[]string{"project", "init"}` (assert exit 0 + stdout == `TODO: init!\n`), and `[]string{"does-not-exist"}` (assert non-zero exit + stderr non-empty). Uses only stdlib `os/exec`; no network
- [X] T028 [P] [US4] In `internal/cli/signal_test.go`, add `TestNotifyContext_CancelsOnSignal` — call `cli.NotifyContext(context.Background())`, send `syscall.SIGTERM` to `os.Getpid()` via `syscall.Kill`, assert returned context's `Done()` channel closes within a short timeout

**Checkpoint**: `go test ./...` runs fully offline in < 30s and breaks loudly when the CLI surface regresses (covers FR-020 and SC-005).

---

## Phase 7: User Story 5 - Continuous Integration on PRs and Main (Priority: P2)

**Goal**: GitHub Actions workflow runs `make test` and `make lint` on
every PR against `main` and every push to `main`.

**Independent Test**: Open a PR with a deliberately failing test — the CI
check fails. Push a fix — the check turns green. Merge to `main` — the
workflow runs once more on the resulting commit.

### Implementation for User Story 5

- [X] T029 [US5] Create `.github/workflows/ci.yml` with name `CI`, triggers `pull_request: branches: [main]` and `push: branches: [main]`, a single `test-and-lint` job on `ubuntu-latest`
- [X] T030 [US5] In `.github/workflows/ci.yml`, add steps: `actions/checkout@v4`, `actions/setup-go@v5` with `go-version-file: go.mod` and module/build caching enabled
- [X] T031 [US5] In `.github/workflows/ci.yml`, add a `Install golangci-lint` step (`golangci/golangci-lint-action@v6` pinned to a recent stable version)
- [X] T032 [US5] In `.github/workflows/ci.yml`, add two observable steps: `name: Test` → `run: make test`, and `name: Lint` → `run: make lint`

**Checkpoint**: A PR against `main` triggers the workflow; both `Test` and `Lint` steps appear as separate observable checks.

---

## Phase 8: Polish & Cross-Cutting Concerns

- [X] T033 [P] Add a top-level `README.md` documenting `make help`, `inject --help`, and the optional `source <(inject _carapace <shell>)` line for shell completion
- [X] T034 Run `make tidy` to ensure `go.mod` / `go.sum` are minimal, then `make lint` and `make test` to confirm a clean tree
- [X] T035 Run the quickstart end-to-end (per `specs/001-cli-scaffold/quickstart.md`) on macOS or Linux and check off each expected outcome

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: no dependencies
- **Foundational (Phase 2)**: depends on Phase 1; blocks every user-story phase
- **US1 (Phase 3)**, **US2 (Phase 4)**: each depends only on Phase 2; can run in parallel by different contributors
- **US3 (Phase 5)**: depends on Phase 2 (needs `./cmd/inject` to exist so `make build` works); independent of US1/US2 implementation details
- **US4 (Phase 6)**: depends on US1, US2, and US3 — the smoke test builds the binary via `make build` (or `go build`) and asserts US1/US2 behavior
- **US5 (Phase 7)**: depends on US3 (calls `make test`/`make lint`) and on US4 (so `make test` has tests to run)
- **Polish (Phase 8)**: depends on everything above

### Within each user story

- Tests are authored alongside implementation; CI guards the end state
- Files marked [P] are independent and can be parallelized

### Parallel opportunities

- Phase 1: T003, T004 in parallel
- Phase 2: T008, T009 in parallel after T005–T007 are in
- Phase 3 tests (T010, T011) in parallel
- Phase 4 tests (T014, T015, T016) in parallel
- Phase 6 tests (T027, T028) in parallel — different files
- Different developers can own US1, US2, US3 in parallel after Phase 2

---

## Parallel Example: User Story 2

```bash
# Author all three US2 tests in parallel (different files / different test funcs):
Task: "TestProjectHelp + TestProjectHelpFlag in internal/cli/project_test.go"
Task: "TestProjectInit_Output in internal/cli/project_init_test.go"
Task: "TestProjectInitHelpFlag in internal/cli/project_init_test.go"
```

---

## Implementation Strategy

### MVP scope (recommended)

1. Phase 1 (Setup)
2. Phase 2 (Foundational)
3. Phase 3 (US1: Discoverable CLI)
4. Phase 4 (US2: `project init` stub)
5. Phase 5 (US3: Makefile)
6. Phase 6 (US4: tests runnable via `make test`)

That is the MVP — it satisfies every P1 story in the spec. CI (US5,
Phase 7) is P2 and can land in a follow-up PR if the MVP is time-boxed.

### Incremental delivery

1. Setup + Foundational → Foundation ready
2. US1 → demoable help screen
3. US2 → demoable `project init` stub
4. US3 → `make build`/`make test`/`make lint` available
5. US4 → tests guard the CLI surface
6. US5 → CI enforces tests + lint on every PR
7. Polish (README, quickstart validation)

---

## Notes

- [P] = different files, no dependency on incomplete tasks
- Every CLI test uses kong's in-process `Parse`/`Run` against captured
  buffers — no subprocess (except the one smoke test in T027)
- Constitution VI (offline tests): T027 builds with `go build` and execs
  the local binary; no network
- Commit per task or per logical group; stop at any checkpoint
