# Implementation Plan: CLI Scaffold & Build Pipeline

**Branch**: `001-cli-scaffold` | **Date**: 2026-05-29 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-cli-scaffold/spec.md`

## Summary

Scaffold the `inject` Go CLI with a hierarchical subcommand tree (`inject`,
`inject project`, `inject project init`), a self-documenting Makefile
(`build`/`test`/`lint`/`release`/`help`), an automated `go test` suite that
exercises the CLI surface, and a GitHub Actions workflow that runs
`make test` and `make lint` on every PR against `main` and every push to
`main`.

The CLI is built on three libraries chosen by the user:

- **`github.com/alecthomas/kong`** — declarative struct-tag-based command parser
  with first-class subcommand groups and auto-generated help text.
- **`github.com/charmbracelet/bubbletea`** — TUI framework reserved for future
  interactive flows; wired into `main` so subcommands can return a
  `tea.Model` without re-plumbing.
- **`github.com/carapace-sh/carapace`** — shell-completion generator that
  reflects on the kong grammar.

Process lifecycle: `main` constructs a `context.Context` derived from
`signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` and passes it
down to every subcommand's `Run` method via a kong bind. On Ctrl-C the
context is cancelled, subcommands return promptly, and `main` exits with
the conventional `130` (SIGINT) status code.

Release/publish is handled by **GoReleaser** invoked from `make release`,
producing macOS and Linux (amd64 + arm64) archives attached to the
corresponding Git tag.

## Technical Context

**Language/Version**: Go 1.26.3 (from `go.mod`)

**Primary Dependencies**:
- `github.com/alecthomas/kong` (CLI parser)
- `github.com/charmbracelet/bubbletea` (TUI runtime, wired but unused by this slice)
- `github.com/carapace-sh/carapace` (shell completion)

**Storage**: N/A (scaffold only; no persisted data)

**Testing**: `go test ./...`; CLI behavior verified via `os/exec` against the
built `./bin/inject` binary inside `internal/cli/*_test.go` and a top-level
`cmd/inject/main_test.go`. No network calls.

**Target Platform**: macOS and Linux (amd64 + arm64). Windows out of scope.

**Project Type**: Single-binary CLI (Go module already initialized as
`github.com/buildoutinc/injector`).

**Performance Goals**: Help screen renders < 50ms cold start. Full `go test`
suite completes < 30s locally, < 2min in CI (SC-004).

**Constraints**:
- No network calls during `go test` (Constitution VI).
- Ctrl-C MUST cancel the root context and propagate to all subcommands;
  process exits with code 130 on SIGINT.
- Self-documenting `make help` driven by inline `##` comments on targets
  (FR-016).

**Scale/Scope**: Single binary, one command group (`project`) with one
subcommand (`init`). Designed to grow to ~10 command groups without
restructuring.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Pluggable Backend Architecture | ✅ N/A | Scaffold only; no backend code in this slice. Plan reserves `internal/backend/` for future backends. |
| II. Inheritance Hierarchy & Multi-Tenancy | ✅ N/A | No domain model introduced. `project` command group is the scaffolding hook. |
| III. Git-Centric Persistence & Schema Validation | ✅ N/A | No persistence introduced. |
| IV. Security & Least Privilege | ✅ PASS | `init` stub does not provision infra yet; no secrets handled; no privileged operations performed. |
| V. Developer Experience & CLI Design | ✅ PASS | Git-like subcommand tree, `--help` on every level (kong default), TUI runtime wired (bubbletea), completion via carapace. |
| VI. Testing & Quality — Local Isolation | ✅ PASS | `go test ./...` runs entirely offline; no AWS calls in this slice. CI invokes `make test` and `make lint`. |

**Tech-stack gate**: Constitution §Technology Stack says CLI framework SHOULD
be "a well-maintained Go CLI library (e.g., `cobra`/`viper`)". `cobra` is
listed as an example, not a mandate. `alecthomas/kong` is a
well-maintained alternative with a declarative struct-tag API that fits
the scaffold's style and the user's explicit instruction. Documented in
Complexity Tracking below.

**Result**: All gates pass.

## Project Structure

### Documentation (this feature)

```text
specs/001-cli-scaffold/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (N/A marker; no domain entities)
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── cli-commands.md  # CLI command-surface contract
└── tasks.md             # (created later by /speckit-tasks)
```

### Source Code (repository root)

```text
cmd/
└── inject/
    └── main.go              # Entrypoint: signal-aware context + kong.Parse + Run

internal/
└── cli/
    ├── root.go              # type RootCmd struct { Project ProjectCmd `cmd:""` ; … }
    ├── root_test.go         # Tests for root help, unknown subcommand exit code
    ├── project.go           # type ProjectCmd struct { Init ProjectInitCmd `cmd:""` }
    ├── project_test.go      # Tests for `inject project` help
    ├── project_init.go      # ProjectInitCmd.Run(ctx context.Context) -> "TODO: init!"
    ├── project_init_test.go # Tests for `inject project init` output + exit code
    ├── completion.go        # carapace registration (`inject _carapace …`)
    └── signal.go            # signal.NotifyContext wrapper (testable)

Makefile                     # self-documenting; build/test/lint/release/help
.github/
└── workflows/
    └── ci.yml               # make test + make lint on PR + push to main
.goreleaser.yaml             # release config consumed by `make release`
go.mod                       # existing
go.sum                       # generated after `go mod tidy`
```

**Structure Decision**: Single-module Go layout. Entrypoint at
`cmd/inject/main.go` (Go convention for binaries shipped from a module that
may grow); all command implementations live in `internal/cli/` so they're
not importable from outside the module. Tests sit next to their source
files (`*_test.go`). No `pkg/` directory until external consumers are
introduced.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Use of `kong` instead of the `cobra` example named in Constitution §Technology Stack | User explicitly selected `kong` for its declarative struct-tag grammar and cleaner subcommand definition; `kong` is well-maintained (the constitution requires only that the chosen library satisfy that property). | `cobra` works but requires more boilerplate per command and a separate `viper` wiring for binding flags; `kong` collapses both into struct tags and produces the same user-facing behavior. |
| `bubbletea` runtime imported by the scaffold without yet rendering a TUI | Required so future interactive subcommands (per Constitution V) can return a `tea.Model` without a follow-up dependency change; cost is one import. | Deferring the dependency would force every future TUI command to add wiring; adding it now is a one-line cost that future-proofs DX. |
