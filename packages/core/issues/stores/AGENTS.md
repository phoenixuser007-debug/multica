<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# stores

## Purpose
Shared stores core module that packages reusable domain logic for the monorepo.

## Key Files
| File | Description |
| --- | --- |
| `draft-store.ts` | Source file local to this directory. |
| `index.ts` | Barrel export for the directory or package. |
| `issues-scope-store.ts` | Source file local to this directory. |
| `my-issues-view-store.ts` | Source file local to this directory. |
| `recent-issues-store.ts` | Source file local to this directory. |
| `selection-store.ts` | Source file local to this directory. |
| `view-store-context.tsx` | Source file local to this directory. |
| `view-store.ts` | Source file local to this directory. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Keep store actions deterministic and avoid wiring route/navigation concerns into store modules.
- Read other stores imperatively only when that dependency direction is already established.

### Testing Requirements
- Run `pnpm typecheck` for TypeScript changes that affect shared types or cross-workspace imports.
- Run the closest Vitest target, or `pnpm test`, when changing UI behavior, hooks, or shared state.

### Common Patterns
- Model state updates as explicit actions and selectors.
- Avoid pushing view-specific formatting or navigation into the store layer.

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
