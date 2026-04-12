<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# core

## Purpose
Shared application core with API access, state, hooks, and domain utilities.

## Key Files
| File | Description |
| --- | --- |
| `eslint.config.mjs` | ESLint configuration for this workspace. |
| `hooks.tsx` | Shared hook exports for the package. |
| `index.ts` | Barrel export for the directory or package. |
| `logger.ts` | Logging helpers used by this package. |
| `package.json` | Package manifest with workspace name, scripts, and dependencies. |
| `provider.tsx` | Provider composition used to wire shared client context. |
| `query-client.ts` | Shared TanStack Query client configuration. |
| `tsconfig.json` | TypeScript configuration for this workspace. |
| `utils.ts` | Source file local to this directory. |
| `vitest.config.ts` | Vitest configuration for local unit tests. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `api/` | Shared api core module that packages reusable domain logic for the monorepo. See `api/AGENTS.md`. |
| `auth/` | Shared auth core module that packages reusable domain logic for the monorepo. See `auth/AGENTS.md`. |
| `chat/` | Shared chat core module that packages reusable domain logic for the monorepo. See `chat/AGENTS.md`. |
| `constants/` | Shared constants core module that packages reusable domain logic for the monorepo. See `constants/AGENTS.md`. |
| `hooks/` | Shared hooks core module that packages reusable domain logic for the monorepo. See `hooks/AGENTS.md`. |
| `inbox/` | Shared inbox core module that packages reusable domain logic for the monorepo. See `inbox/AGENTS.md`. |
| `issues/` | Shared issues core module that packages reusable domain logic for the monorepo. See `issues/AGENTS.md`. |
| `modals/` | Shared modals core module that packages reusable domain logic for the monorepo. See `modals/AGENTS.md`. |
| `navigation/` | Shared navigation core module that packages reusable domain logic for the monorepo. See `navigation/AGENTS.md`. |
| `pins/` | Shared pins core module that packages reusable domain logic for the monorepo. See `pins/AGENTS.md`. |
| `platform/` | Shared platform core module that packages reusable domain logic for the monorepo. See `platform/AGENTS.md`. |
| `projects/` | Shared projects core module that packages reusable domain logic for the monorepo. See `projects/AGENTS.md`. |
| `realtime/` | Shared realtime core module that packages reusable domain logic for the monorepo. See `realtime/AGENTS.md`. |
| `runtimes/` | Shared runtimes core module that packages reusable domain logic for the monorepo. See `runtimes/AGENTS.md`. |
| `types/` | Shared types core module that packages reusable domain logic for the monorepo. See `types/AGENTS.md`. |
| `workspace/` | Shared workspace core module that packages reusable domain logic for the monorepo. See `workspace/AGENTS.md`. |

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
- `zustand` via `packages/core/package.json`.
- `@tanstack/react-query` via `packages/core/package.json`.
- `vitest` via `packages/core/package.json`.
- `typescript` via `packages/core/package.json`.
- `@multica/tsconfig` via `packages/core/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
