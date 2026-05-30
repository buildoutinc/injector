# Phase 1 Data Model: CLI Scaffold & Build Pipeline

**Feature**: 001-cli-scaffold | **Date**: 2026-05-29

## Scope

This feature is **scaffold-only**. It introduces no domain entities,
no persisted state, and no schemas. The spec explicitly notes this in its
"Key Entities" section.

## Code-level structures (not domain entities)

For completeness, the kong command-grammar structs that this scaffold
introduces are listed here — these are *parser inputs*, not data models,
and live in `internal/cli/`:

| Struct | Purpose | File |
|--------|---------|------|
| `RootCmd` | Top-level CLI; declares `Project` subcommand and global flags | `internal/cli/root.go` |
| `ProjectCmd` | `inject project` command group; declares `Init` subcommand | `internal/cli/project.go` |
| `ProjectInitCmd` | `inject project init`; `Run(ctx) error` prints `TODO: init!` | `internal/cli/project_init.go` |

State transitions, persistence, and validation rules: **N/A** for this slice.

Real domain entities (Organization, Project, Service, Environment, Schema)
arrive in later features and will be modeled per Constitution Principles
II and III.
