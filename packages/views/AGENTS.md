<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# views

## Purpose
Feature-level view modules composed from core state and shared UI primitives.

## Key Files
| File | Description |
| --- | --- |
| `eslint.config.mjs` | ESLint configuration for this workspace. |
| `package.json` | Package manifest with workspace name, scripts, and dependencies. |
| `tsconfig.json` | TypeScript configuration for this workspace. |
| `vitest.config.ts` | Vitest configuration for local unit tests. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `agents/` | Agents view module containing UI composition for that feature area. See `agents/AGENTS.md`. |
| `auth/` | Auth view module containing UI composition for that feature area. See `auth/AGENTS.md`. |
| `chat/` | Chat view module containing UI composition for that feature area. See `chat/AGENTS.md`. |
| `common/` | Common view module containing UI composition for that feature area. See `common/AGENTS.md`. |
| `editor/` | Editor view module containing UI composition for that feature area. See `editor/AGENTS.md`. |
| `inbox/` | Inbox view module containing UI composition for that feature area. See `inbox/AGENTS.md`. |
| `issues/` | Issues view module containing UI composition for that feature area. See `issues/AGENTS.md`. |
| `layout/` | Layout view module containing UI composition for that feature area. See `layout/AGENTS.md`. |
| `modals/` | Modals view module containing UI composition for that feature area. See `modals/AGENTS.md`. |
| `my-issues/` | My issues view module containing UI composition for that feature area. See `my-issues/AGENTS.md`. |
| `navigation/` | Navigation view module containing UI composition for that feature area. See `navigation/AGENTS.md`. |
| `projects/` | Projects view module containing UI composition for that feature area. See `projects/AGENTS.md`. |
| `runtimes/` | Runtimes view module containing UI composition for that feature area. See `runtimes/AGENTS.md`. |
| `search/` | Search view module containing UI composition for that feature area. See `search/AGENTS.md`. |
| `settings/` | Settings view module containing UI composition for that feature area. See `settings/AGENTS.md`. |
| `skills/` | Skills view module containing UI composition for that feature area. See `skills/AGENTS.md`. |
| `test/` | Shared test utilities for the views package. See `test/AGENTS.md`. |
| `workspace/` | Workspace view module containing UI composition for that feature area. See `workspace/AGENTS.md`. |

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
- See the parent directory guidance in `../AGENTS.md` for adjacent modules that work with this area.

### External
- `vitest` via `packages/views/package.json`.
- `typescript` via `packages/views/package.json`.
- `@base-ui/react` via `packages/views/package.json`.
- `@dnd-kit/core` via `packages/views/package.json`.
- `@dnd-kit/sortable` via `packages/views/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
