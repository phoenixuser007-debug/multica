<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# hooks

## Purpose
Shared hooks core module that packages reusable domain logic for the monorepo.

## Key Files
| File | Description |
| --- | --- |
| `use-file-upload.ts` | Source file local to this directory. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Keep hooks focused on one concern and isolate side effects from derived state.
- Expose a minimal API so calling components stay easy to test.

### Testing Requirements
- Run `pnpm typecheck` for TypeScript changes that affect shared types or cross-workspace imports.
- Run the closest Vitest target, or `pnpm test`, when changing UI behavior, hooks, or shared state.

### Common Patterns
- Separate derived state from side effects so hooks stay testable.
- Return a narrow surface area to keep call sites readable.

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
