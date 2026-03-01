# Decision 0001: Flat YAML config schema

**Date:** 2026-03-01
**Status:** Accepted
**Feature:** /specs/03-app-registry.md

## Context

The spec example uses a nested YAML schema (`orchestratr.leader_key`, `orchestratr.chord_timeout_ms`) while the existing example config at `configs/example.yml` uses flat top-level keys (`leader_key`, `api_port`). We needed to choose one canonical format before implementing the registry.

## Options

1. **Flat top-level keys** — simpler to parse, matches existing `configs/example.yml`, fewer indentation levels for users editing by hand
2. **Nested under `orchestratr:` namespace** — cleaner if the config file is ever shared with other tools, but adds unnecessary nesting for a single-purpose config

## Decision

Use flat top-level keys. orchestratr is a standalone tool with its own config directory; namespacing within the file adds complexity with no benefit. The existing example config already uses this format. We also include `ready_cmd` and `ready_timeout_ms` as optional per-app fields (forward-compatible with spec 05, cross-env launch).

## Consequences

- Spec 03 example config section should be updated to match the flat schema
- All code references use flat struct fields (no nested `Orchestratr` wrapper)
- If a multi-tool config is ever needed (unlikely per non-goals), migration will be required
