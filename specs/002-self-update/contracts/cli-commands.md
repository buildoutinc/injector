# CLI Contract: Self-Update Surface

**Feature**: 002-self-update | **Date**: 2026-05-30

User-facing contract for the CLI changes this feature lands. Test code
asserts against this exactly.

## New / changed command tree

```text
inject
├── --help, -h, help            # root help (US2 notice may append)
├── --version, -v               # one-line "inject <version>" + newline
├── version                     # NEW — detailed version (US2 notice may append)
├── upgrade                     # NEW — self-update flow
│   ├── --check                 # report-only; no disk writes
│   └── --pre-release           # include pre-release tags
├── project
│   └── init
└── _carapace <shell>           # unchanged
```

## Environment variables

| Variable | Effect |
|----------|--------|
| `INJECT_NO_UPDATE_CHECK` | Any non-empty value disables the passive update check **and** any inline notice. Honored by every command. |
| `GITHUB_TOKEN` | Optional; passed to the selfupdate library to lift unauthenticated rate limits. Never required for end users. |

## `inject version`

- **Exit code**: 0
- **Stdout**:

  ```text
  inject <semver-or-"dev"> (commit <short-sha>, built <yyyy-mm-dd>)
  ```

  The `commit` and `built` fields MAY be absent on local `make build`
  output; the semver field is always present.
- **Stderr**: empty
- **Passive notice**: rendered to stdout below the version line when
  (a) `INJECT_NO_UPDATE_CHECK` is unset, (b) the cached or freshly
  fetched latest release is strictly newer than the running binary,
  and (c) the running binary version is not `"dev"`.
- **Maps to**: FR-009, FR-013, FR-014

## `inject` and `inject help` and `inject --help` and `inject -h`

- **Exit code**: 0
- **Stdout**: the help screen (unchanged from feature 001) followed
  by the same passive notice block, when applicable.
- **Stderr**: empty
- **Maps to**: FR-009, FR-013

## `inject upgrade`

- **Exit code**: 0 on success or "already up to date"; non-zero on any
  failure path (network, checksum mismatch, permission denied,
  unsupported platform).
- **Stdout** on upgrade: `inject: upgraded vA.B.C → vX.Y.Z\n`
- **Stdout** on no-op: `inject: already up to date (vX.Y.Z)\n`
- **Stderr** on failure: one human-readable sentence naming the cause.
  Examples:
  - `inject: cannot upgrade: download failed: <reason>` (network)
  - `inject: cannot upgrade: checksum mismatch — release tampered or transfer corrupt` (FR-004)
  - `inject: cannot upgrade: /opt/inject is not writable by your user — install via your package manager (e.g. brew upgrade injector)` (FR-006)
  - `inject: cannot upgrade: no release artifact for <os>/<arch>` (SC-003 / unsupported platform)
- **Maps to**: FR-001, FR-002, FR-003, FR-004, FR-006, FR-008

### `inject upgrade --check`

- **Exit code**: 0 always (it's a query).
- **Stdout**:
  - `A newer version is available: vX.Y.Z (you have vA.B.C)` when newer.
  - `Already on the latest version (vX.Y.Z)` when up to date.
- **No disk writes; no cache writes.**
- **Maps to**: FR-005

### `inject upgrade --pre-release`

Same as `inject upgrade` but considers pre-release tags. Combinable
with `--check`. **Maps to**: FR-007.

## Passive-notice render contract

When applicable, exactly this block is appended to stdout, preceded by
one blank line:

```text

==> A newer version is available: vX.Y.Z
==> Run "inject upgrade" to install it.
```

- TTY stdout: `==>` is cyan.
- Non-TTY stdout: plain text (no ANSI sequences).
- Appears at most once per process invocation, even if multiple
  trigger points exist.
- **Maps to**: FR-009, FR-013

## Failure & guard cases

| Scenario | Behavior |
|----------|----------|
| Running on an unsupported OS/arch | `inject upgrade` → exit non-zero with platform message; passive notice → never runs on this build (the goroutine returns early) |
| `INJECT_NO_UPDATE_CHECK=1` set | No upstream call; no notice; `inject upgrade --check` still works because the user asked for it explicitly |
| Cache file corrupt | Treated as cache miss; refreshed silently |
| Network unreachable | `inject upgrade` returns non-zero; passive notice falls through silently (FR-011) |
| Binary version is `dev` | Passive notice skipped entirely (research §8) |
