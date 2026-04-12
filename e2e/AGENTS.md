<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# e2e

## Purpose
Playwright end-to-end tests and shared test fixtures for user workflows.

## Key Files
| File | Description |
| --- | --- |
| `auth.spec.ts` | End-to-end or integration test covering this flow. |
| `comments.spec.ts` | End-to-end or integration test covering this flow. |
| `fixtures.ts` | Source file local to this directory. |
| `helpers.ts` | Source file local to this directory. |
| `issues.spec.ts` | End-to-end or integration test covering this flow. |
| `navigation.spec.ts` | End-to-end or integration test covering this flow. |
| `settings.spec.ts` | End-to-end or integration test covering this flow. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Prefer API setup and cleanup helpers over brittle UI bootstrapping in tests.
- Keep specs independent so they can run in isolation and in parallel.

### Testing Requirements
- Run `pnpm exec playwright test <spec>` for the affected flows after UI or API behavior changes.
- Keep fixtures self-contained and reuse `e2e/helpers.ts` and `e2e/fixtures.ts` where possible.

### Common Patterns
- Follow the closest existing local pattern before inventing a new one.
- Prefer small, scoped edits so this directory stays easy to navigate.

## Dependencies

### Internal
- Targets the running backend and web app through browser-level flows.

### External
- `@playwright/test` via `package.json`.
- `typescript` via `package.json`.
- `@types/node` via `package.json`.
- `@types/pg` via `package.json`.
- `pg` via `package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
