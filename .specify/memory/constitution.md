<!--
SYNC IMPACT REPORT
==================
Version change: N/A (first population) → 1.0.0
Version bump rationale: MINOR — initial content population from blank template; all sections new.

Modified principles: (none — all sections are new)

Added sections:
  - Core Principles (I–VI)
  - Technology Stack
  - Personas & Target Environments
  - Governance

Removed sections: (none)

Templates requiring updates:
  - .specify/templates/plan-template.md  ✅ No changes required; Constitution Check section
    is generic and references the constitution file at runtime.
  - .specify/templates/spec-template.md  ✅ No changes required; structure aligns with
    Injector's functional requirement and user-story model.
  - .specify/templates/tasks-template.md ✅ No changes required; phase/story structure
    matches Go project conventions described here.

Follow-up TODOs:
  - TODO(RATIFICATION_DATE): Treat 2026-05-28 as ratification date (first committed date).
    Update if the team formally adopts a different date.
  - TODO(ORG_INFRA): Organization-level IAM policies and SSM path conventions not yet
    specified. Refine in a future amendment once first-backend design is complete.
-->

# Injector Constitution

## Core Principles

### I. Pluggable Backend Architecture

The backend layer MUST be pluggable. Backends are exclusively and solely responsible
for encryption, storage, and access control. No business logic, schema validation, or
inheritance resolution may be embedded in a backend implementation.

The first and reference backend implementation is **AWS SSM Parameter Store**. New
backends MUST implement the same interface without requiring changes to CLI or core
business logic.

### II. Inheritance Hierarchy & Multi-Tenancy

Injector MUST model a strict inheritance hierarchy:
**Organization → Project → Service → Environment**.

- All projects inherit from their Organization; all environments inherit from their
  Project (which itself inherits from the Organization).
- Individual secrets and variables MAY be overridden at any lower level.
- A project MAY define any number of environments. Default environments are
  `development`, `staging`, and `production`. Users MAY add or remove environments.
- The tool MUST natively support multiple applications (projects) per organization and
  multiple services per application.

### III. Git-Centric Persistence & Schema Validation

All project data and configuration MUST be persisted as plain text files (YAML).
The format MUST be optimized for Git workflows, human-readable diffs, and code review.
Secrets values MUST never be stored in plain text in tracked files.

A schema MUST define the structure and validation rules for secrets and variables.
Schemas are applied at the service level and apply to all service environments.
Schemas MUST:
- Validate the type and presence (required/optional) of every secret and variable on
  write.
- Be usable to validate an entire environment on demand.
- Support type-casting when loading values into a process environment.
- Be versioned in Git and require explicit approval to change (enforced via CI check).

### IV. Security & Least Privilege

The tool MUST adhere to the principle of least privilege at every level:
- Provisioned IAM roles and policies grant only the permissions required for the
  declared operation.
- Secret values MUST NOT be visible in CLI output unless the user explicitly requests
  them and has the required access.
- The `init` command MUST scaffold the minimum required infrastructure
  (SSM Parameter Store paths, IAM roles, policies) without over-provisioning.
- Engineer Interns MUST be restricted to the `development` environment and MUST NOT
  be able to view secret values or mutate any other environment.

Security best practices MUST be the path of least resistance — the tool makes doing
the right thing easy and doing the wrong thing hard.

### V. Developer Experience & CLI Design

The CLI MUST follow a git-like subcommand structure: intuitive, hierarchical, and
self-documenting. Every command MUST include excellent built-in help text (`--help`).

- Provide a rich TUI with colored, well-formatted output on modern terminals.
- Lean heavily on the user's `$EDITOR` for complex or bulk edits (clean YAML blobs).
- Support one-line `init` to scaffold a new project and provision required
  infrastructure.
- Support importing existing `.env` files with a frictionless workflow.
- Prefer convention over configuration throughout.

### VI. Testing & Quality — Local Isolation (NON-NEGOTIABLE)

The test suite MUST be capable of running entirely locally with no external network
calls during automated testing.

- Unit and integration tests MUST use a local mock of AWS SSM Parameter Store
  (prefer a free/open-source alternative to LocalStack; e.g., a lightweight in-memory
  SSM stub).
- End-to-end (e2e) tests MAY target a real AWS environment but MUST be gated behind
  an explicit build tag or environment variable so they never run automatically in CI
  without explicit opt-in.
- All new features MUST be accompanied by unit tests. Schema validation and
  inheritance resolution logic MUST have integration-level coverage.

## Technology Stack

- **Language**: Go (Golang). All code MUST be idiomatic, clean, and robust Go.
- **Data Formats**: YAML for all bulk data, configuration schemas, and editor
  integrations.
- **CLI Framework**: Use a well-maintained Go CLI library (e.g., `cobra`/`viper`).
- **Testing**: `go test` for unit and integration; build-tagged e2e suite for AWS.
- **Mocking**: In-process SSM stub for integration tests (no Docker daemon required
  for `go test ./...`).

## Personas & Target Environments

**Target Environments**: Local development (Linux, macOS), Docker containers,
CI/CD pipelines (GitHub Actions), AWS ECS, and EC2.

**Personas**:

- **Team Engineer**: Creates a new project, scaffolds project files, and provisions
  required infrastructure via `injector init`.
- **Engineer**: CRUDs secrets/variables for application environments, manages schemas,
  sets up ECS/EC2 deployments. Reviews schema PRs and unblocks failing CI checks.
- **Platform Engineer**: Manages the organization's applications and services. Runs
  GitHub Actions checks in CI/CD to block deployments until all required secrets are
  set in every environment.
- **Engineer Intern**: Read-only access to schema and development environment.
  Can propose schema changes (YAML PR) but cannot view secret values or mutate
  non-development environments. Changes require engineer approval and a passing CI
  schema-validation check before merge.

## Governance

This Constitution supersedes all other practices, style guides, and prior conventions
for the Injector project. All architectural decisions, code generation, and test
writing MUST comply with it.

**Amendment procedure**:
1. Propose amendment as a PR modifying this file.
2. At least one senior maintainer MUST review and approve.
3. A migration plan MUST be included if the amendment is a MAJOR (backward-incompatible)
   change.
4. Update `LAST_AMENDED_DATE` and increment `CONSTITUTION_VERSION` per semver rules
   (MAJOR / MINOR / PATCH as described in version policy).

**Version policy**:
- MAJOR: Backward-incompatible principle removal or redefinition.
- MINOR: New principle, section added, or materially expanded guidance.
- PATCH: Clarifications, wording fixes, non-semantic refinements.

**Compliance review**: Every PR MUST verify compliance with the Constitution Check
gate in the implementation plan (`plan.md`) before merging.

**Version**: 1.0.0 | **Ratified**: 2026-05-28 | **Last Amended**: 2026-05-28
