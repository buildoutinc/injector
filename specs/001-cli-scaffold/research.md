# Phase 0 Research: CLI Scaffold & Build Pipeline

**Feature**: 001-cli-scaffold | **Date**: 2026-05-29

All NEEDS CLARIFICATION items resolved. Each section documents one decision.

---

## 1. CLI parser: `alecthomas/kong`

**Decision**: Use `github.com/alecthomas/kong` as the CLI parser.

**Rationale**:
- Declarative grammar via struct tags: subcommand tree, flags, args, and
  help text live next to the type that implements `Run(...)`.
- Per-command `Run(deps...) error` signature lets kong inject a
  `context.Context` (and any other shared dependency) via `kong.Bind`,
  matching the user's "pass a cancelable context to the main program"
  requirement cleanly.
- Generates a `--help` screen for every node in the command tree out of the
  box, satisfying FR-002, FR-003, FR-004, FR-007.
- Unknown subcommands cause `kong.Parse` to print usage to stderr and exit
  non-zero — satisfies FR-008 without custom code.
- Well-maintained (active commits, broadly adopted in the Go ecosystem).

**Alternatives considered**:
- `spf13/cobra` — works but requires more boilerplate per command and
  separate `viper` wiring for flag binding. The constitution lists it as
  an *example*, not a mandate.
- `urfave/cli/v3` — fewer struct-tag conveniences; less ergonomic
  context binding.

---

## 2. TUI runtime: `charmbracelet/bubbletea`

**Decision**: Import `github.com/charmbracelet/bubbletea` in `main.go`
(unused this slice; reserved for future interactive subcommands).

**Rationale**:
- Constitution V mandates a "rich TUI with colored, well-formatted output".
- Wiring bubbletea now is one import; deferring it would mean retrofitting
  every future interactive command.
- Bubbletea's `tea.Program` accepts a `context.Context` via
  `tea.WithContext`, so the same cancelable context that fronts the CLI
  also cancels any future TUI.

**Alternatives considered**:
- `rivo/tview` — heavier, more "widget-shaped", less ergonomic for the
  Elm-style update loops the team will likely want.
- Deferring entirely — rejected because future TUI commands would each pay
  the dependency-add cost.

---

## 3. Shell completion: `carapace-sh/carapace`

**Decision**: Use `github.com/carapace-sh/carapace` for shell completion.

**Rationale**:
- Generates completion for bash, zsh, fish, elvish, and PowerShell from a
  single registration.
- Has a documented integration story with kong (reflect over the kong
  application to produce the carapace spec).
- Active fork of the original `rsteube/carapace`, which is now archived;
  `carapace-sh` is the maintained continuation.

**Implementation note**: Completion is registered behind a hidden
`_carapace` subcommand; user opt-in is documented in the README via
`source <(inject _carapace <shell>)`. Not part of FR list — included
because the user explicitly chose the library.

**Alternatives considered**:
- `posener/complete` — older, single-shell focus.
- Hand-rolled bash completion — rejected; doesn't scale to multiple
  shells.

---

## 4. Signal handling & cancelable context

**Decision**: Use `signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)`
in `main` and pass the resulting `ctx` into every subcommand via
`kong.Bind(ctx)`.

**Rationale**:
- `signal.NotifyContext` (stdlib, available since Go 1.16) is the canonical
  pattern; no third-party dependency needed.
- On Ctrl-C the returned `stop` func is invoked and the context is
  cancelled — subcommands return promptly when they observe `ctx.Done()`.
- Exit code on SIGINT is set to `128 + 2 = 130` to follow POSIX
  conventions. Implementation: `main` checks `errors.Is(err,
  context.Canceled)` after `kong.Run` returns and exits 130.

**Alternatives considered**:
- `oklog/run` — overkill for a single signal source.
- Manual `os/signal.Notify` + `select` — works but `signal.NotifyContext`
  is one line and gives a real `context.Context` for free.

---

## 5. Self-documenting Makefile

**Decision**: Targets are documented inline with a `## description` comment
on the target line. `make help` parses the Makefile with `awk` to print
two aligned columns.

**Rationale**:
- Single source of truth: the description sits on the same line as the
  target, so adding a target without a description is visually obvious in
  PRs.
- No external tool required (`awk` is POSIX).
- Satisfies FR-016 (new targets discovered automatically).

**Template snippet** (illustrative):

```makefile
help:  ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
```

**Alternatives considered**:
- A separate `HELP.md` — rejected, drifts from reality.
- `mage` — adds a Go-based build tool; team is asking for a plain
  Makefile.

---

## 6. Release tooling: GoReleaser

**Decision**: Use **GoReleaser** invoked from `make release`. Config lives in
`.goreleaser.yaml`. CI does not run release; releases are cut by a
maintainer locally or by pushing a tag (a follow-up release workflow
can be added when needed).

**Rationale**:
- The de-facto standard for Go binary releases. Handles multi-OS/arch
  builds, archive creation, checksums, and GitHub Releases uploads in one
  invocation.
- Spec §Assumptions explicitly cites GoReleaser as the expected tool.
- Satisfies FR-015 (macOS + Linux, amd64 minimum) with a single config.

**Alternatives considered**:
- Hand-rolled `go build` + `gh release create` — works but reimplements
  what GoReleaser already does well.

---

## 7. Linter: `golangci-lint`

**Decision**: Use `golangci-lint` invoked from `make lint`.

**Rationale**:
- Community standard for Go (spec §Assumptions calls it out by name).
- Aggregates multiple linters (`govet`, `staticcheck`, `errcheck`,
  `ineffassign`, `unused`) behind one binary.
- Config in `.golangci.yaml` keeps rules versioned with the repo.

**Alternatives considered**:
- `staticcheck` alone — narrower coverage.
- `revive` — additional config burden; less industry adoption.

---

## 8. CI: GitHub Actions matrix

**Decision**: Single Linux job (`ubuntu-latest`) running `make test` and
`make lint` on PRs to `main` and pushes to `main`. macOS coverage
deferred (FR-019 makes macOS optional).

**Rationale**:
- Minimal compute footprint; fastest signal.
- Spec §SC-006 requires < 5 min feedback — Linux-only keeps total time
  well under that.

**Alternatives considered**:
- Matrix on `{ubuntu, macos}` — useful but doubles CI cost for a release
  that doesn't require it yet.

---

## 9. Testing strategy for the CLI surface

**Decision**: Two layers of `go test`:

1. **Unit tests** on `internal/cli/*` that invoke the kong parser
   in-process (no subprocess) and capture stdout/stderr via
   `bytes.Buffer`. Covers FR-002, FR-003, FR-004, FR-006, FR-007, FR-008.
2. **One smoke test** in `cmd/inject` that builds the binary via
   `go test -run` + `os/exec` to confirm the wired-up binary behaves
   identically end-to-end. Skipped under `testing.Short()`.

**Rationale**:
- In-process tests are fast (< 30s total, SC-004) and offline (Constitution
  VI).
- One subprocess smoke test catches wiring regressions (e.g., missing
  `kong.Bind`, broken signal handling) that in-process tests miss.

**Alternatives considered**:
- Pure subprocess tests — slower, harder to assert on internal state.
- Pure in-process tests — miss wiring bugs.
