# Phase 0 Research: Self-Update via GitHub Releases

**Feature**: 002-self-update | **Date**: 2026-05-30

All decisions documented. No NEEDS CLARIFICATION items outstanding.

---

## 1. Self-update library: `rhysd/go-github-selfupdate`

**Decision**: Use `github.com/rhysd/go-github-selfupdate/selfupdate`.

**Rationale**:
- Spec-mandated.
- Active project, broadly used. Built around GitHub Releases.
- Has the exact features the spec needs: latest-release detection,
  semver comparison, multi-format archive support (`.zip`, `.gzip`,
  `.tar.gz`, `.tar.xz`, uncompressed), atomic binary replacement with
  rollback, optional hash & signature validation, pre-release opt-in,
  and `GITHUB_TOKEN` support to lift rate limits.

**Library API choice**: `selfupdate.UpdateSelf(currentVersion, slug)`
is the one-shot helper. We use the lower-level `selfupdate.Updater`
(constructed via `NewUpdater(Config{...})`) so we can:
- pass `Filters` / `Validator` to enforce checksum validation (FR-004),
- pass `Prerelease: true` only when `--pre-release` is set (FR-007),
- pass an HTTP client with a timeout so `--check` is bounded,
- pass a `Source` other than the default GitHub one **in tests** (the
  library exposes a `Source` interface used by `UpdaterSource`).

**Alternatives considered**:
- `creativeprojects/go-selfupdate` — fork with multi-host support
  (GitHub/Gitea/GitLab). Not selected; spec mandates `rhysd`.
- `minio/selfupdate` — lower level, no GitHub-aware release discovery;
  too DIY for this feature.

---

## 2. Artifact naming compatibility

**Decision**: Keep the existing GoReleaser archive name template
`{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}` from feature
001 (e.g., `inject_0.2.0_linux_amd64.tar.gz`).

**Rationale**:
- go-github-selfupdate's matcher uses a tolerant regex that accepts
  `_` or `-` as separators and only requires `…{os}{sep}{arch}…` to
  appear in the asset name; the embedded version between `{cmd}` and
  `{os}` does not break matching.
- Library docs (Naming Rules of Released Binaries) explicitly support
  `.tar.gz`. Confirmed by the library's `UncompressCommand` helper
  which dispatches on URL extension.
- Reusing the existing template avoids a churn change to feature 001's
  release config; the new release.yml just runs the same goreleaser
  invocation, plus a `checksums.txt` (already configured).

**Note**: The library also requires the *inside* of the archive to
contain an executable named exactly `inject`. GoReleaser's default
puts the binary at the root of the archive named after the build's
`binary:` field, which we set to `inject`. ✅

---

## 3. Permission pre-check (FR-006)

**Decision**: On Unix, call `unix.Access(binPath, unix.W_OK)` (via
`golang.org/x/sys/unix`). If it returns `EACCES`, `EROFS`, or any
other error, refuse the upgrade with a message that names the path
and suggests `brew upgrade injector` (or the user's package manager).

**Rationale**:
- `os.Stat()` + mode-bit math is wrong on systems with ACLs.
- `unix.Access` mirrors what `access(2)` does and reflects the
  effective UID/GID — the correct question for "can I write here".

**Alternatives considered**:
- Attempt the write and recover — leaves a half-written tempfile in
  the bin directory; bad UX and bad for read-only mounts.

---

## 4. 24h update-check cache

**Decision**: Cache file path: `os.UserCacheDir()` + `/inject/update-check.json`.
JSON shape:

```json
{
  "checked_at": "2026-05-30T14:22:00Z",
  "latest_version": "v0.3.0",
  "binary_version": "v0.1.0"
}
```

- `checked_at` < 24h ago → no network call; render notice from
  `latest_version` if `latest_version > binary_version`.
- `checked_at` ≥ 24h ago → async network call refreshes the file.
- `binary_version` is stored so a `go install`-installed dev binary
  doesn't keep nagging about an irrelevant cached result after the
  user updates manually.

**Rationale**:
- `os.UserCacheDir()` honors `XDG_CACHE_HOME` on Linux and resolves to
  `~/Library/Caches` on macOS — exactly what the spec's assumptions
  call for.
- One small file, easy to inspect during debugging (SC-005's
  verification path).

**Concurrency**: Two `inject` invocations racing on the cache file is
handled by `os.WriteFile` (atomic on POSIX for small writes via
`tmp+rename`); the worst outcome is one of the two writes being
overwritten, which is harmless.

---

## 5. Async update check & 200ms budget (SC-004)

**Decision**: When the foreground command starts, spawn one goroutine
that calls `updatecheck.Check(ctx)` with `ctx, cancel :=
context.WithTimeout(parent, 200*time.Millisecond)`. The foreground
command does its work and then, just before returning, calls
`select { case n := <-noticeCh: render(n); case <-ctx.Done(): }`.

**Rationale**:
- Hides network latency behind work the foreground command does
  anyway (parsing args, rendering help, reading version metadata).
- If the foreground command finishes before the check (or vice
  versa), the slower one drops out; the budget can never be exceeded
  because `ctx.Done()` is the floor.
- A cache hit returns in microseconds, so the common case is "notice
  always shows when it should."

**Alternatives considered**:
- Synchronous check with tight timeout — measurable slowdown on every
  invocation when the cache is warm, because the cache check itself
  is fast but adding a network attempt at all has a worse worst case.

---

## 6. Changelog: `git-chglog` with Conventional Commits

**Decision**: Use `git-chglog` (`github.com/git-chglog/git-chglog`) in
the release workflow. Config under `.chglog/`:

- `config.yml` declares the commit pattern (Conventional Commits) and
  groups commits into named sections (`Features`, `Fixes`,
  `Performance`, `Refactor`, `Documentation`, `Other Changes`).
- `CHANGELOG.tpl.md` renders the section into Markdown that GoReleaser
  attaches as the release body via `release.header_file` or
  `release.body_file`.
- Commits whose subjects don't match any Conventional Commit type fall
  through to a catch-all template clause (`Other Changes`), satisfying
  FR-017.

**Rationale**:
- GoReleaser's built-in changelog supports grouping but doesn't surface
  un-prefixed commits in a named section; un-prefixed commits would be
  silently dropped from the user-visible sections, violating FR-017.
- `git-chglog` config is small (one YAML + one template) and runs in
  the workflow with no Go-side code.

**Alternatives considered**:
- `goreleaser changelog` — see above.
- `release-please` — more opinionated; would also tag versions for us;
  rejected because the spec assumes humans push tags (Assumptions §4).

---

## 7. Release workflow trigger & fork safety (FR-015)

**Decision**: `release.yml` triggers only on `push` of `refs/tags/v*.*.*`,
with `if: github.repository == 'buildoutinc/injector'` on every job.

**Rationale**:
- `pull_request` and `pull_request_target` are never triggers, so a
  fork can't run the release pipeline.
- The repo-name guard is belt-and-suspenders for fork pushes via
  `workflow_dispatch` triggered from forks.
- GitHub doesn't propagate secrets to forked PRs by default, but
  guarding by tag-push event keeps the surface minimal.

**Idempotency** (FR-021): GoReleaser's `--clean` flag wipes the
`dist/` dir each run, and uploading the same asset twice fails (GitHub
returns 422). The workflow treats that as a loud failure (job fails)
rather than catching the error — this is the spec's "fail loudly" path.

---

## 8. Version embedding (FR-014, FR-018)

**Decision**: Build with `-ldflags "-X main.version={{.Version}}"`.
`main.go` declares `var version = "dev"`; GoReleaser overrides it at
link time. `inject version` reads `cli.Version()` which returns this
value. For local `make build`, version stays `"dev"`.

**Rationale**:
- Standard Go pattern, supported natively by GoReleaser's `builds.ldflags`.
- No code generation or runtime-file dependencies.
- The "dev" sentinel makes the passive check on local builds always
  silent because no remote release can be "newer than dev" without
  ambiguity — we'll treat `dev` as "skip the check entirely" in
  `updatecheck.Check`.

---

## 9. Notice format & stdout vs stderr (FR-013)

**Decision**: Print the notice to **stdout** after the foreground
command's normal output, separated by a blank line and prefixed with
`==>` (cyan when stdout is a TTY, plain otherwise). Example:

```text
inject v0.1.0 (commit abc1234, built 2026-05-30)

==> A newer version is available: v0.3.0
==> Run "inject upgrade" to install it.
```

**Rationale**:
- FR-013 says "at the end of stdout"; this honors that.
- The `==>` marker + leading blank line keeps machine-parseable
  output (the version line) unchanged — automation that reads the
  first line of stdout is unaffected.
- Detect TTY via `term.IsTerminal(int(os.Stdout.Fd()))` for color;
  same library bubbletea already pulls in.
