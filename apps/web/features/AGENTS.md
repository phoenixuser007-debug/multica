<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# features

## Purpose
App-local feature modules that have not been extracted into shared packages.

## Key Files
No direct source files live here; this directory mainly organizes subdirectories.

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `auth/` | Auth-specific client logic that is local to the web app. See `auth/AGENTS.md`. |
| `landing/` | Landing-page specific presentation and localization logic. See `landing/AGENTS.md`. |

## For AI Agents

### Working In This Directory
- Keep changes scoped to this directory and follow existing naming and layering patterns.
- Prefer extending existing modules over introducing parallel abstractions without a clear need.

### Testing Requirements
- Run `pnpm typecheck` for TypeScript changes that affect shared types or cross-workspace imports.
- Run the closest Vitest target, or `pnpm test`, when changing UI behavior, hooks, or shared state.

### Common Patterns
- Follow the closest existing local pattern before inventing a new one.
- Prefer small, scoped edits so this directory stays easy to navigate.

## Dependencies

### Internal
- `apps/web/components/` and `packages/ui/` provide shared UI composition primitives.

### External
- `next` via `apps/web/package.json`.
- `react` via `apps/web/package.json`.
- `zustand` via `apps/web/package.json`.
- `@tanstack/react-query` via `apps/web/package.json`.
- `vitest` via `apps/web/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
