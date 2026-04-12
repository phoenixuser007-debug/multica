<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# hooks

## Purpose
Hooks view module containing UI composition for that feature area.

## Key Files
| File | Description |
| --- | --- |
| `index.ts` | Barrel export for the directory or package. |
| `use-issue-reactions.ts` | Source file local to this directory. |
| `use-issue-subscribers.ts` | Source file local to this directory. |
| `use-issue-timeline.ts` | Source file local to this directory. |

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
- `packages/core/` provides domain state and API access used by views.
- `packages/ui/` provides shared presentational primitives.

### External
- `vitest` via `packages/views/package.json`.
- `typescript` via `packages/views/package.json`.
- `@base-ui/react` via `packages/views/package.json`.
- `@dnd-kit/core` via `packages/views/package.json`.
- `@dnd-kit/sortable` via `packages/views/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
