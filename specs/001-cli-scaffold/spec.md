# Feature Specification: CLI Scaffold & Build Pipeline

**Feature Branch**: `001-cli-scaffold`

**Created**: 2026-05-28

**Status**: Draft

**Input**: User description: "Create a `inject` CLI command that when run outputs a help screen with a summary of sub-commands. Also create an `inject project init` subcommand that for now outputs \"TODO: init!\" to STDOUT for now. Include a Makefile that is self documenting and includes goals to build, test, lint and release / publish `inject` to Github Releases. Finally include a test suite that can be run from `make` and a Github Actions workflow to run test and lint when a PR is opened or commits pushed to main."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Discoverable CLI Entry Point (Priority: P1)

A new contributor (or end user) runs the `inject` CLI for the first time and needs to
understand what the tool does and what it can do. Running the binary with no
arguments — or with `--help` / `-h` — must produce a clear help screen that lists
every available subcommand with a one-line summary, plus global flags.

**Why this priority**: This is the first impression of the entire tool. Without a
discoverable entry point, every other feature is invisible. It is also the foundation
on which all future subcommands hang. Per the project constitution, the CLI MUST be
self-documenting (Principle V), so this story is non-negotiable for the first release.

**Independent Test**: Build the binary, run `inject`, `inject --help`, and `inject -h`.
Each must exit successfully (exit code 0) and print a help screen containing the tool
name, a short description, the list of subcommands (currently just `project`), and
global flags. Delivers immediate user value as a working, browsable CLI surface.

**Acceptance Scenarios**:

1. **Given** a freshly built `inject` binary on macOS or Linux, **When** the user runs
   `inject` with no arguments, **Then** the help screen is printed to standard output
   and the process exits with code 0.
2. **Given** a freshly built `inject` binary, **When** the user runs `inject --help`
   or `inject -h`, **Then** the same help screen is printed and the process exits
   with code 0.
3. **Given** the help screen output, **When** the user reads it, **Then** they see
   the tool name (`inject`), a one-line description, a "Commands" section listing
   each registered subcommand with a short summary, and a "Flags" section listing
   global flags.
4. **Given** the user runs an unknown subcommand (`inject does-not-exist`), **When**
   the CLI dispatches, **Then** an error is printed to standard error, the help
   screen is shown, and the process exits with a non-zero exit code.

---

### User Story 2 - Stubbed `project init` Subcommand (Priority: P1)

A user runs `inject project init` to scaffold a new Injector project. For this first
slice the subcommand is a placeholder that prints `TODO: init!` to standard output and
exits successfully. This establishes the subcommand tree (`inject project init`),
proves the dispatch wiring, and gives downstream work a real command to flesh out.

**Why this priority**: This is the minimum proof that the subcommand hierarchy works
end-to-end (root → `project` → `init`). Without it, Story 1's help screen would list
no real commands, and downstream stories (real `init` behavior) would have no
attachment point. Required for MVP.

**Independent Test**: Build the binary, run `inject project init`. The process must
print exactly `TODO: init!` (followed by a newline) to standard output and exit with
code 0. Also confirm `inject project` (no further args) and `inject project --help`
print a `project`-scoped help screen listing `init` as a subcommand.

**Acceptance Scenarios**:

1. **Given** a built `inject` binary, **When** the user runs `inject project init`,
   **Then** standard output contains `TODO: init!` followed by a newline, standard
   error is empty, and exit code is 0.
2. **Given** a built `inject` binary, **When** the user runs `inject project`,
   **Then** a help screen scoped to the `project` command is printed, listing `init`
   as an available subcommand, and exit code is 0.
3. **Given** a built `inject` binary, **When** the user runs `inject project init
   --help`, **Then** help text specific to `init` is printed and exit code is 0.

---

### User Story 3 - Self-Documenting Build & Release Makefile (Priority: P1)

A developer cloning the repository for the first time wants to know how to build,
test, lint, and publish the tool without reading source code or onboarding docs.
Running `make` (or `make help`) at the repo root must print a list of every available
make target with a short description. The Makefile must include, at minimum, targets
to **build** the `inject` binary, **run tests**, **run lint**, and **release/publish**
the binary to GitHub Releases.

**Why this priority**: This is the developer-facing entry point and is required for
Stories 4 and 5 (the test suite is run via `make test`; CI invokes `make test` and
`make lint`). Without it, the build/test/release loop is undefined.

**Independent Test**: Clone a fresh checkout, run `make` with no arguments. Output
must be a help screen listing each target and its description. Then run `make build`
and verify the `inject` binary is produced; run `make test` and verify the test suite
executes; run `make lint` and verify a linter executes against the codebase.

**Acceptance Scenarios**:

1. **Given** a fresh repo checkout, **When** the developer runs `make` or `make help`,
   **Then** standard output lists each Makefile target along with a human-readable
   description, in a single readable column-aligned format.
2. **Given** a fresh repo checkout with the required toolchain installed, **When**
   the developer runs `make build`, **Then** an `inject` executable is produced in a
   predictable location (e.g., `./bin/inject` or `./dist/inject`) and the process
   exits with code 0.
3. **Given** a fresh repo checkout, **When** the developer runs `make test`, **Then**
   the project's automated test suite executes and the process exits with code 0 if
   all tests pass.
4. **Given** a fresh repo checkout, **When** the developer runs `make lint`, **Then**
   a code linter runs against the codebase and the process exits with code 0 if no
   lint issues are found.
5. **Given** a tagged release commit, **When** an authorized maintainer runs
   `make release` (or the equivalent target name), **Then** a release artifact
   (binary for at least macOS and Linux) is built and published to GitHub Releases,
   attached to the corresponding tag.
6. **Given** any new target is added to the Makefile, **When** a contributor inspects
   it, **Then** the target's description is declared inline (e.g., via a recognized
   comment convention) so that `make help` picks it up automatically without manual
   list maintenance.

---

### User Story 4 - Automated Test Suite Runnable from Make (Priority: P1)

A developer needs to verify their changes do not break behavior before opening a PR.
The repository ships with an automated test suite that exercises the CLI's behavior
(help output, exit codes, `project init` stub output). Running `make test` executes
every test, prints pass/fail results, and exits non-zero on any failure.

**Why this priority**: Tests are the safety net for every story above and below. Per
the constitution (Principle VI), the test suite MUST run locally with no external
network calls.

**Independent Test**: Run `make test` on a clean checkout. All tests pass and exit
code is 0. Then intentionally break a behavior covered by tests (e.g., change the
`TODO: init!` output) and re-run; at least one test must fail and exit code must be
non-zero.

**Acceptance Scenarios**:

1. **Given** a clean checkout with the project's declared toolchain, **When** the
   developer runs `make test`, **Then** every test in the suite executes locally
   with no network calls, results are printed to standard output, and exit code is 0
   on success.
2. **Given** the `project init` subcommand is broken (e.g., wrong output), **When**
   the developer runs `make test`, **Then** the relevant test fails, the failure is
   clearly reported, and exit code is non-zero.
3. **Given** the help command is broken (e.g., missing the `project` subcommand
   listing), **When** the developer runs `make test`, **Then** the relevant test
   fails and exit code is non-zero.

---

### User Story 5 - Continuous Integration on PRs and Main (Priority: P2)

A maintainer wants every pull request and every push to `main` to be automatically
checked for test and lint regressions before merging or deploying. A GitHub Actions
workflow runs the test suite and the linter on every PR opened against the main
branch and on every commit pushed to `main`. The workflow status appears as a
required check on PRs.

**Why this priority**: Important for team workflows and quality gates but downstream
of having tests and a working Makefile (Stories 3, 4). Without CI the gate is
manual; with it the project's bar is enforced automatically.

**Independent Test**: Open a PR against `main` containing a deliberately failing
test. The GitHub Actions workflow runs and reports a failing status check on the PR.
Push a passing commit to the PR branch; the workflow re-runs and reports success.
Merge to `main`; the workflow runs again on the resulting commit.

**Acceptance Scenarios**:

1. **Given** a PR is opened against `main`, **When** GitHub processes the PR event,
   **Then** the CI workflow executes `make test` and `make lint` and reports a
   pass/fail status check on the PR.
2. **Given** a commit is pushed to `main`, **When** GitHub processes the push event,
   **Then** the CI workflow executes `make test` and `make lint` and reports
   pass/fail status on the commit.
3. **Given** the CI workflow runs, **When** any step fails, **Then** the overall
   workflow status is reported as failed and the PR's merge button reflects the
   failing required check (per repository branch protection settings).
4. **Given** the CI workflow runs, **When** all steps succeed, **Then** the overall
   workflow status is reported as success.

---

### Edge Cases

- **Unknown subcommand**: `inject some-typo` MUST print an error to standard error,
  show the help screen, and exit with a non-zero exit code (covered by US1 scenario 4).
- **Help while inside a subcommand**: `inject project init --help` MUST print
  init-specific help, not the root help.
- **Make target with no description**: A new target added without a description
  comment SHOULD still be executable but MAY be omitted from `make help` output;
  this is acceptable as long as documented targets are complete.
- **Release on non-tagged commit**: Running `make release` from a non-tagged commit
  SHOULD either fail with a clear error message or publish a pre-release/snapshot —
  the choice is documented in `make help`.
- **CI on forks**: PRs from forks may not have access to release credentials; the
  test and lint jobs MUST still run on fork PRs even if release-related secrets are
  unavailable.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST be invokable as a single executable named `inject`.
- **FR-002**: Running `inject` with no arguments MUST print a help screen to standard
  output and exit with code 0.
- **FR-003**: Running `inject --help` or `inject -h` MUST print the same help screen
  as `inject` with no arguments and exit with code 0.
- **FR-004**: The help screen MUST include the tool name, a one-line description, a
  "Commands" section listing every registered subcommand with a one-line summary,
  and a "Flags" section listing global flags.
- **FR-005**: The CLI MUST register a `project` command group with at least one
  subcommand: `init`.
- **FR-006**: Running `inject project init` MUST print exactly `TODO: init!` followed
  by a newline to standard output and exit with code 0.
- **FR-007**: Running `inject project` with no further arguments or with `--help` MUST
  print a help screen scoped to `project` that lists `init` as an available
  subcommand.
- **FR-008**: Running `inject` with an unknown subcommand MUST print an error to
  standard error, show the help screen, and exit with a non-zero exit code.
- **FR-009**: The repository MUST include a `Makefile` at the repository root.
- **FR-010**: Running `make` with no arguments or `make help` MUST print a
  self-documenting list of every available target with a short description.
- **FR-011**: The Makefile MUST include at minimum these targets: `build`, `test`,
  `lint`, `release` (or equivalent name for "publish to GitHub Releases"), and
  `help`.
- **FR-012**: `make build` MUST produce an `inject` executable in a predictable
  output location.
- **FR-013**: `make test` MUST execute the project's full test suite locally with no
  external network calls and exit non-zero on any failure.
- **FR-014**: `make lint` MUST execute a code linter against the codebase and exit
  non-zero on any reported issue.
- **FR-015**: `make release` MUST build release artifacts for at least macOS and
  Linux (amd64 at minimum) and publish them to GitHub Releases attached to the
  corresponding git tag.
- **FR-016**: New Makefile targets MUST follow a convention (e.g., inline doc
  comment) that allows `make help` to discover and list them automatically without
  manual maintenance of a help list.
- **FR-017**: The repository MUST include a GitHub Actions workflow that runs on
  every pull request opened against `main` and on every commit pushed to `main`.
- **FR-018**: The CI workflow MUST execute `make test` and `make lint` as separate,
  observable steps and report overall pass/fail status to GitHub.
- **FR-019**: The CI workflow MUST run on at least one Linux runner; macOS coverage
  MAY be added but is not required for the first release.
- **FR-020**: The test suite MUST include automated tests that verify FR-002, FR-003,
  FR-004, FR-006, FR-007, and FR-008.

### Key Entities *(include if feature involves data)*

This feature is scaffold-only and introduces no domain data entities. The notable
artifacts are:

- **`inject` binary**: The compiled CLI executable; the user-facing surface.
- **Makefile**: The developer-facing entry point; declares build/test/lint/release
  targets and a self-documenting `help`.
- **CI Workflow**: A declarative pipeline that runs on PR and push-to-main events,
  invoking `make test` and `make lint`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new contributor can build the binary, see the help screen, and run
  `inject project init` in under 5 minutes from a fresh clone, with no documentation
  beyond `make help` and `inject --help`.
- **SC-002**: 100% of supported entry points (`inject`, `inject --help`, `inject -h`,
  `inject project`, `inject project --help`, `inject project init`, `inject project
  init --help`) exit with the documented exit code on a clean build.
- **SC-003**: 100% of declared Makefile targets are listed in `make help` output and
  every listed target executes successfully on a clean checkout.
- **SC-004**: The test suite completes in under 30 seconds on a developer laptop and
  in under 2 minutes on the CI runner.
- **SC-005**: At least one test exists for each of FR-002, FR-003, FR-004, FR-006,
  FR-007, and FR-008, and intentionally breaking the corresponding behavior causes
  at least one test to fail.
- **SC-006**: Every PR opened against `main` displays a passing or failing CI status
  check within 5 minutes of being opened or updated (assuming the runner is
  available).
- **SC-007**: A maintainer can cut a release by pushing a git tag and running a
  single `make release` invocation, with the resulting artifacts published and
  downloadable from GitHub Releases within 10 minutes.

## Assumptions

- The tool is distributed as a single statically linked native executable; users do
  not need a separate runtime to run `inject`.
- The repository targets macOS and Linux as primary developer platforms; Windows
  support is out of scope for this first release.
- The `release` target uses an authenticated GitHub credential (e.g., a token
  supplied via environment variable or CI secret) to upload artifacts; setting that
  credential up is the maintainer's responsibility and not in scope for the spec.
- GitHub Actions is the CI provider; no other CI provider is targeted in this
  release.
- The linter selection follows community convention for the chosen implementation
  language (e.g., `golangci-lint` for Go); the spec does not mandate a specific
  linter, only that one is run via `make lint`.
- "Publish to GitHub Releases" uses a standard release tool appropriate for the
  ecosystem (e.g., `goreleaser` for Go). The spec does not mandate the tool, only
  the outcome.
- Branch protection rules requiring CI to pass before merge are a repository-admin
  concern; the spec only guarantees the workflow exists and reports status.
