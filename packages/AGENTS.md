<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# packages

## Purpose
Shared packages consumed by the application workspaces.

## Key Files
No direct source files live here; this directory mainly organizes subdirectories.

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `core/` | Shared application core with API access, state, hooks, and domain utilities. See `core/AGENTS.md`. |
| `eslint-config/` | Reusable ESLint presets shared across the monorepo. See `eslint-config/AGENTS.md`. |
| `tsconfig/` | Reusable TypeScript base configs shared across workspaces. See `tsconfig/AGENTS.md`. |
| `ui/` | Shared UI primitives, styles, and markdown helpers. See `ui/AGENTS.md`. |
| `views/` | Feature-level view modules composed from core state and shared UI primitives. See `views/AGENTS.md`. |

## For AI Agents

### Working In This Directory
- Keep changes scoped to this directory and follow existing naming and layering patterns.
- Prefer extending existing modules over introducing parallel abstractions without a clear need.

### Testing Requirements
- Run the closest relevant automated check for files changed in this directory.
- If there is no dedicated automation, validate imports, paths, and rendered output manually.

### Common Patterns
- Follow the closest existing local pattern before inventing a new one.
- Prefer small, scoped edits so this directory stays easy to navigate.

## Dependencies

### Internal
- See the root `AGENTS.md` for top-level repository relationships.

### External
- `@playwright/test` via `package.json`.
- `typescript` via `package.json`.
- `@types/node` via `package.json`.
- `@types/pg` via `package.json`.
- `pg` via `package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
