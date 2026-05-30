# CLI Command-Surface Contract

**Feature**: 001-cli-scaffold | **Date**: 2026-05-29

This is the user-facing contract for the `inject` binary as of this slice.
It is the source of truth that the test suite verifies. Any change to
this file MUST be reflected in tests and vice versa.

## Binary

- **Name**: `inject`
- **Distribution**: single statically linked native executable
- **Supported platforms (release artifacts)**: `darwin/amd64`,
  `darwin/arm64`, `linux/amd64`, `linux/arm64`

## Global behavior

- Process responds to **SIGINT** (Ctrl-C) and **SIGTERM** by cancelling its
  root `context.Context` and exiting with code **130** (SIGINT) or **143**
  (SIGTERM).
- Help is rendered to **stdout** on success paths and to **stderr** on
  error paths.
- All commands accept `--help` / `-h` and exit 0 after printing help.

## Command tree

```text
inject                       # root help
├── --help, -h                # global flag
├── --version                 # global flag, prints version + commit + date
└── project                   # command group
    └── init                  # subcommand
```

## Contracts

### `inject` (no args) and `inject --help` / `inject -h`

- **Exit code**: 0
- **Stdout**: help screen containing:
  - Tool name: `inject`
  - One-line description
  - `Commands:` section listing `project` with a one-line summary
  - `Flags:` section listing `--help`, `--version`
- **Stderr**: empty
- **Maps to**: FR-002, FR-003, FR-004

### `inject <unknown-subcommand>`

- **Exit code**: non-zero (kong default: 1)
- **Stdout**: empty
- **Stderr**: error message naming the unknown command + usage hint
- **Maps to**: FR-008

### `inject project` and `inject project --help`

- **Exit code**: 0
- **Stdout**: help screen scoped to `project`, listing `init` as a
  subcommand
- **Stderr**: empty
- **Maps to**: FR-005, FR-007

### `inject project init`

- **Exit code**: 0
- **Stdout**: exactly `TODO: init!\n` (the literal string `TODO: init!`
  followed by a single LF newline)
- **Stderr**: empty
- **Maps to**: FR-006

### `inject project init --help`

- **Exit code**: 0
- **Stdout**: help screen scoped to `init`
- **Stderr**: empty
- **Maps to**: FR-007 (extension)

### Signal handling

- **SIGINT received during any command**: subcommand's `Run(ctx)` observes
  `ctx.Done()`, returns promptly; process exits with code 130.
- **SIGTERM received during any command**: same as SIGINT but exit code 143.
- For the `TODO: init!` stub, the subcommand completes before any signal
  could meaningfully interrupt it; the contract becomes load-bearing once
  long-running subcommands are added.

## Makefile contract

`make` and `make help` MUST print one line per documented target. Targets
declared by this slice:

| Target | Description |
|--------|-------------|
| `help` | Show this help. |
| `build` | Build `./bin/inject` for the host platform. |
| `test` | Run the full Go test suite locally. |
| `lint` | Run `golangci-lint` against the codebase. |
| `release` | Build cross-platform archives and publish to GitHub Releases via GoReleaser. |
| `tidy` | Run `go mod tidy`. |
| `clean` | Remove `./bin` and other build artifacts. |

## CI workflow contract

`.github/workflows/ci.yml`:

- **Triggers**: `pull_request` against `main`, `push` to `main`.
- **Runner**: `ubuntu-latest`.
- **Steps** (separate, observable):
  1. Check out repo
  2. Set up Go (version from `go.mod`)
  3. `make test`
  4. `make lint`
- **Status**: overall pass/fail reported to GitHub PR checks.
- **Maps to**: FR-017, FR-018, FR-019.
