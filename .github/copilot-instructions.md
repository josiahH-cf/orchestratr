# Copilot Instructions

## Project Standards

- Read `/AGENTS.md` before starting any task — it defines the project's conventions, testing rules, commit practices, and review expectations

## Completions

- Match naming conventions and patterns in the file being edited
- Prefer explicit types over inferred when the language supports both
- Do not generate placeholder or TODO comments

## Code Review

- Flag functions over 50 lines
- Flag nesting deeper than 3 levels
- Flag missing error handling on I/O operations
- Flag tests that assert only the happy path
- Flag hardcoded values that should be configuration
- Verify every acceptance criterion from the linked spec has a test

## PR Descriptions

- State what changed, why, and how to verify
- Link to the spec in /specs/ if one exists
- List files changed, grouped by concern

## Coding Agent

- Read `/AGENTS.md` before starting any task — it defines the feature lifecycle and agent boundaries
- For new features: create a spec in `/specs/` using `/specs/_TEMPLATE.md`, then a task breakdown in `/tasks/` using `/tasks/_TEMPLATE.md`
- Do not create GitHub issues, labels, or milestones — these are managed by humans
- Do not modify files outside the scope described in the spec
- Follow the feature lifecycle: Scope → Plan → Test → Implement → Review
- Read the linked spec before writing any code
- One task per session; commit after each task passes its tests
