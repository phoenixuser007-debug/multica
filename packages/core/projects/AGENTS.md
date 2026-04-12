<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# projects

## Purpose
Shared projects core module that packages reusable domain logic for the monorepo.

## Key Files
| File | Description |
| --- | --- |
| `config.ts` | Source file local to this directory. |
| `index.ts` | Barrel export for the directory or package. |
| `mutations.ts` | Source file local to this directory. |
| `queries.ts` | Source file local to this directory. |

## Subdirectories
No documented subdirectories.

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
- Sibling `packages/core/*` modules provide adjacent domain state and helpers.

### External
- `zustand` via `packages/core/package.json`.
- `@tanstack/react-query` via `packages/core/package.json`.
- `vitest` via `packages/core/package.json`.
- `typescript` via `packages/core/package.json`.
- `@multica/tsconfig` via `packages/core/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
