# Feature Specification: Self-Update via GitHub Releases

**Feature Branch**: `002-self-update`

**Created**: 2026-05-30

**Status**: Draft

**Input**: User description: "Implement self-updating from a GitHub release for
`github.com/buildoutinc/injector`. Add a tagging+release workflow that produces
release notes from semantic commits and ships compressed Linux/macOS binaries
named in the format the self-update library expects. Add an `inject upgrade`
subcommand that installs the latest release if one is newer than the running
binary. Have the CLI passively check for new releases and inform the user when
they run `inject version` or see the root help (`inject` / `inject help`).
Document the release and update process in `README.md`."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - One-Command Self-Upgrade (Priority: P1)

A user who already has `inject` installed wants the newest version without
visiting a web page, downloading an archive, unpacking it, replacing the
binary, or restoring file permissions. They run `inject upgrade`. The tool
discovers the latest published release, decides whether it is newer than
what is currently running, downloads the right artifact for their OS and
CPU, verifies it, swaps the running binary atomically, and prints a clear
summary of what just happened.

**Why this priority**: This is the entire point of the feature — users
having a frictionless path to the newest version. Without it the rest of
the feature (release pipeline, version checks) has no payoff. Required for
MVP.

**Independent Test**: From a machine running a pinned older `inject`
binary, run `inject upgrade`. The process exits 0, prints the
before/after version and the release tag installed, and the binary on disk
reports the new version when re-run. If the running version is already the
latest, the same command exits 0 and prints "already up to date" without
overwriting the binary.

**Acceptance Scenarios**:

1. **Given** an installed `inject` binary at version `v0.1.0` and a
   published release `v0.2.0`, **When** the user runs `inject upgrade`,
   **Then** the binary on disk is replaced atomically with the v0.2.0
   release artifact for the user's OS/arch, stdout shows
   `inject: upgraded v0.1.0 → v0.2.0` (or equivalent), and exit code is 0.
2. **Given** an installed `inject` binary already at the latest release,
   **When** the user runs `inject upgrade`, **Then** stdout reads
   `inject: already up to date (v0.X.Y)`, the binary is not modified, and
   exit code is 0.
3. **Given** the user is offline, **When** they run `inject upgrade`,
   **Then** stderr explains the network failure in one sentence, the
   binary is not modified, and exit code is non-zero.
4. **Given** the user has installed `inject` via a package manager (e.g.,
   `brew`) or to a path they do not own, **When** they run
   `inject upgrade`, **Then** the tool detects the permission problem
   before downloading, prints a clear remediation message that tells them
   to use their package manager instead, and exits non-zero.
5. **Given** the user passes `inject upgrade --check`, **When** the
   command runs, **Then** the tool only reports whether an upgrade is
   available (and to what version) without downloading or modifying
   anything, and exit code is 0.

---

### User Story 2 - Passive "Newer Version Available" Notice (Priority: P1)

A user runs `inject version` or sees the root help (`inject` or
`inject help`) — both natural moments to learn about the tool's current
state. If a newer release exists upstream, a one-line, non-intrusive
notice appears at the bottom of the output telling them the latest
version and how to upgrade (`inject upgrade`). The check is rate-limited
so it doesn't run on every invocation, never blocks the foreground
command, and never causes the command to fail on network errors.

**Why this priority**: Users will not discover the upgrade command on
their own; surfacing it where they already are converts "feature exists"
into "feature gets used". This is what makes the upgrade pathway
discoverable and used. Required for MVP.

**Independent Test**: On a fresh machine running a pinned older version,
run `inject version`. The version line is printed; immediately after, a
notice appears: `A newer version is available: vX.Y.Z (you have vA.B.C).
Run "inject upgrade" to install it.` Repeat the command within the
rate-limit window and confirm the network is not hit again (no notice
update). Disable the network and confirm the command still prints the
version and exits 0 with no notice.

**Acceptance Scenarios**:

1. **Given** the rate-limit window has elapsed since the last check and
   a newer release exists, **When** the user runs `inject version`,
   **Then** the version line prints to stdout followed by a one-line
   upgrade notice naming the newer version and the `inject upgrade`
   command, and exit code is 0.
2. **Given** the running version is already the latest, **When** the
   user runs `inject version`, **Then** the version line prints to
   stdout and no upgrade notice is shown, and exit code is 0.
3. **Given** the rate-limit window has not elapsed since the last
   check, **When** the user runs `inject version` or `inject` (root
   help), **Then** the foreground command does not perform a network
   call and the cached result (if any) is used for the notice.
4. **Given** the user is offline or the upstream check fails for any
   reason, **When** the user runs `inject version` or `inject`,
   **Then** the foreground command completes normally (exit 0) and
   **no** error or notice about the check failure is printed.
5. **Given** the user has set the opt-out (env var or config flag),
   **When** any command runs, **Then** no upstream check is performed
   and no upgrade notice ever appears.
6. **Given** `inject` or `inject help` is run, **When** the help screen
   prints, **Then** the upgrade notice (when applicable) appears once,
   at the end of the help output.

---

### User Story 3 - Automated Tagged GitHub Release (Priority: P1)

A maintainer wants to cut a new release without manually writing release
notes, building binaries, or naming artifacts. They push a git tag of the
form `vX.Y.Z` to `main`. A CI workflow takes over: it determines what
commits are new since the previous tag, generates a changelog from
**Conventional Commits** (`feat:`, `fix:`, etc.), compiles the binary for
Linux and macOS on both `amd64` and `arm64`, compresses each archive in
the **exact name format the self-update library expects** so that
`inject upgrade` can find and install it, and publishes everything as a
GitHub Release attached to that tag with the generated notes.

**Why this priority**: User Story 1 (`inject upgrade`) and User Story 2
(passive notice) both depend on actual releases being published in a
format the self-update library can consume. Without this pipeline, the
other two stories have nothing to talk to. Required for MVP.

**Independent Test**: From a clean repo, create a few commits using
Conventional Commit prefixes (`feat:`, `fix:`, `chore:`), push a `vX.Y.Z`
tag, and wait for CI. A GitHub Release appears with the tag, the
release notes group commits by type (Features / Fixes / Other), and the
Release's Assets list contains compressed binaries for `linux_amd64`,
`linux_arm64`, `darwin_amd64`, `darwin_arm64` with names matching the
update library's expected pattern. A local `inject upgrade` against this
release succeeds.

**Acceptance Scenarios**:

1. **Given** a tag `vX.Y.Z` is pushed to the repository, **When** the
   release workflow runs, **Then** a GitHub Release for `vX.Y.Z` is
   created, its notes contain a changelog grouped by Conventional Commit
   type since the previous tag, and its assets include compressed
   binaries for the four supported OS/arch combos plus a checksums file.
2. **Given** the same tag is pushed again or the workflow re-runs, **When**
   the release is processed, **Then** the existing release is not
   silently duplicated; the workflow either updates the existing release
   idempotently or fails loudly with a clear error.
3. **Given** the tag does not match `vX.Y.Z` (semver), **When** the
   workflow runs, **Then** the workflow fails fast with a clear message
   and does not publish a release.
4. **Given** any commit since the previous tag is missing a Conventional
   Commit prefix, **When** notes are generated, **Then** that commit
   is still included under a generic "Other Changes" section so nothing
   is silently dropped.
5. **Given** the artifacts are uploaded, **When** `inject upgrade` runs
   on a developer machine targeting that release, **Then** the upgrade
   succeeds end-to-end (artifact named correctly, downloaded,
   uncompressed, swapped in).

---

### Edge Cases

- **Pre-release tags** (`v1.0.0-rc.1`, `v1.0.0-beta.2`): treated as
  pre-releases on GitHub and **ignored** by the default upgrade and
  notice paths; users who want them must opt in via a flag (e.g.,
  `inject upgrade --pre-release`).
- **Tag pushed by a fork PR**: the release workflow MUST NOT run from
  fork events because release credentials would be exposed; tags pushed
  to `main` are the only trigger.
- **Binary running on an unsupported OS/arch** (e.g., `windows`,
  `linux/386`): `inject upgrade` exits non-zero with a clear "no
  artifact for your platform" message; the passive notice does not run
  in this case.
- **Self-replacement on Windows**: out of scope for v1 (Windows is not
  a supported target).
- **Downgrades**: not supported by `inject upgrade`; the command only
  installs strictly newer semver releases. Explicit downgrade is a
  manual install (documented in README).
- **Two upgrades racing on the same machine**: the second one detects
  the binary is already current and reports "already up to date".
- **Self-check ping cadence**: the cadence and persistence location of
  the rate-limit cache live in the README; they are implementation
  details and not part of the user-visible contract beyond "doesn't
  hit the network on every invocation."
- **Checksums and integrity**: every downloaded artifact MUST be
  integrity-checked against the published checksums file before the
  swap; failure aborts the upgrade and leaves the existing binary
  intact.
- **Permission-denied path**: if the binary lives at a path the user
  cannot overwrite (Homebrew cellar, system path, read-only mount), the
  upgrade MUST refuse before downloading and explain the situation.

## Requirements *(mandatory)*

### Functional Requirements

**Upgrade command (`inject upgrade`)**

- **FR-001**: The CLI MUST provide an `inject upgrade` subcommand,
  discoverable from `inject --help` like every other subcommand.
- **FR-002**: `inject upgrade` MUST look up the latest stable release for
  `github.com/buildoutinc/injector`, compare it semantically to the
  running binary's version, and replace the running binary only if the
  remote version is strictly newer.
- **FR-003**: When already up to date, `inject upgrade` MUST print a
  clear "already up to date" line and exit 0 without writing to disk.
- **FR-004**: When upgrading, `inject upgrade` MUST verify the downloaded
  artifact against the release's published checksums file before swapping
  the running binary; a checksum mismatch MUST abort the upgrade with the
  existing binary intact and exit non-zero.
- **FR-005**: `inject upgrade --check` MUST report whether an upgrade is
  available (and to what version) without modifying anything; exit 0.
- **FR-006**: `inject upgrade` MUST refuse and exit non-zero with a clear
  remediation message when the current binary's path is not writable by
  the current user, telling the user to upgrade via their package
  manager instead.
- **FR-007**: `inject upgrade` MUST support `--pre-release` (or
  equivalent) to opt into pre-release tags; without it, only stable
  semver releases are considered.
- **FR-008**: `inject upgrade` MUST exit non-zero with a one-sentence
  network-failure message when it cannot reach the release source, and
  MUST NOT leave a partial binary on disk.

**Passive new-version notice**

- **FR-009**: When `inject version` is run, `inject` (no args) is run,
  or `inject help` (or `inject --help`) is run, the CLI MUST, after the
  normal output, append at most one line informing the user that a
  newer version is available and how to install it, **only** when a
  newer version actually exists.
- **FR-010**: The check that backs FR-009 MUST be rate-limited; the
  default cadence is documented in the README and MUST be at most one
  network call per 24-hour window per user.
- **FR-011**: The check that backs FR-009 MUST never fail the foreground
  command; on any error (network down, GitHub rate-limited, parse
  failure) the command completes normally with no notice.
- **FR-012**: Users MUST be able to disable the passive check via an
  environment variable (`INJECT_NO_UPDATE_CHECK=1` or equivalent) and/or
  a config flag; when disabled, no upstream check is ever made.
- **FR-013**: The notice MUST appear at most once per command invocation,
  at the end of stdout (or after the help screen), and MUST be visually
  distinct enough that automation parsing the normal output is not
  confused — i.e., the version line stays unchanged.

**Version subcommand**

- **FR-014**: The CLI MUST provide an `inject version` subcommand that
  prints the current version to stdout and exits 0. The format MUST
  include at minimum the semver version of the running binary; commit
  SHA and build date MAY also be included.

**Release pipeline**

- **FR-015**: A GitHub Actions workflow MUST trigger on pushes of tags
  matching `vX.Y.Z` (and `vX.Y.Z-<prerelease>`) on the repository's
  default branch and MUST NOT trigger from forked pull requests.
- **FR-016**: The workflow MUST generate release notes that group the
  commits since the previous tag by Conventional Commit type (at minimum:
  Features, Fixes, Other), preserving commit subjects.
- **FR-017**: Any commit since the previous tag that lacks a Conventional
  Commit prefix MUST still appear in the notes under an "Other Changes"
  (or equivalent) section — none silently dropped.
- **FR-018**: The workflow MUST build the `inject` binary for the
  combinations `linux/amd64`, `linux/arm64`, `darwin/amd64`,
  `darwin/arm64`, MUST embed the release version into the binary so
  `inject version` reports it, and MUST compress each binary into an
  archive whose name matches the format the self-update library expects
  to discover via the GitHub Releases API.
- **FR-019**: The workflow MUST publish a checksums file alongside the
  binaries and MUST upload all artifacts to the GitHub Release for the
  tag.
- **FR-020**: Tags that are not valid semver (per FR-015's pattern) MUST
  NOT produce a release; the workflow fails fast with a clear message.
- **FR-021**: The release workflow MUST be idempotent or fail loudly on
  re-runs of the same tag — never silently duplicate a release.

**Documentation**

- **FR-022**: The repository's `README.md` MUST document:
  (a) how an end user upgrades (`inject upgrade`, `inject upgrade
  --check`, opt-out env var); (b) how a maintainer cuts a release
  (Conventional Commit conventions, tag format, what CI does); (c) what
  artifacts the release publishes and their naming pattern.

### Key Entities

- **Release**: A published GitHub Release for the repository,
  identified by a semver tag (`vX.Y.Z`). Carries: release notes
  (changelog), a set of OS/arch binary archives, and a checksums file.
- **Version**: The semver string baked into the running binary at build
  time and reported by `inject version`. Compared semantically against
  the latest Release to decide upgrade eligibility.
- **Update Check Record**: A small, local, per-user record (timestamp +
  last-known-latest version) that backs FR-010's rate limit. Storage
  location is implementation-defined and documented in the README.
- **Conventional Commit**: A commit whose subject begins with a
  recognized type prefix (`feat:`, `fix:`, `chore:`, `docs:`,
  `refactor:`, `test:`, `build:`, `ci:`, `perf:`, optionally with a
  scope and `!` for breaking changes). Used only by the release
  workflow to group changelog entries.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From a clean machine running an older `inject`, a user
  can run a single command (`inject upgrade`) and end up on the latest
  version in under 30 seconds on a typical broadband connection.
- **SC-002**: After running `inject upgrade` to install the latest
  release, the on-disk binary's `inject version` output matches the
  release tag exactly.
- **SC-003**: 100% of unsupported-platform invocations of
  `inject upgrade` exit non-zero with a single, human-readable message
  that names the user's platform.
- **SC-004**: The passive update check adds no more than 200 ms of
  wall-clock time to `inject version`, `inject`, and `inject help` on a
  warm cache (and never more than that, even on a network failure that
  must be detected and abandoned).
- **SC-005**: The passive update check performs at most one network
  request per 24-hour window per user on a given machine, verifiable by
  inspecting the persisted check record after repeated invocations.
- **SC-006**: When a tag matching `vX.Y.Z` is pushed to `main`, the
  corresponding GitHub Release exists with generated notes and all four
  expected platform archives plus a checksums file, within 10 minutes
  of the tag push on a typical CI run.
- **SC-007**: For any tagged release produced by the workflow, every
  commit between the previous tag and the new tag appears in the
  release notes — none missing, none duplicated.
- **SC-008**: A new contributor can read the README sections on
  upgrading and releasing and successfully perform each flow (install
  the newer binary; cut a release from a tag) without consulting any
  other documentation.

## Assumptions

- The repository is `github.com/buildoutinc/injector`; the self-update
  library is configured to read releases from this exact repo.
- The default branch is `main`; release tags are pushed to `main`.
- The supported target platforms are Linux and macOS on `amd64` and
  `arm64`. Windows and other architectures are out of scope for v1.
- The release process is initiated by a human pushing a tag; the
  workflow does not auto-bump versions or auto-tag from merges.
- Conventional Commits are the team's commit convention, but are
  **not** enforced in this feature beyond the release-notes grouping —
  unprefixed commits are tolerated and surface under "Other Changes".
- GitHub Releases are the only distribution channel for v1. Package
  managers (Homebrew, apt, etc.) may republish but are out of scope.
- The rate-limit cache for the passive check lives under the user's
  standard cache directory (e.g., `$XDG_CACHE_HOME/inject/` on Linux,
  `~/Library/Caches/inject/` on macOS); this is documented in the
  README and not user-configurable beyond opt-out.
- The repository remains public; the self-update library does not need
  authenticated GitHub API access for end users.
