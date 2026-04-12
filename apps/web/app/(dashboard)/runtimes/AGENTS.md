<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# runtimes

## Purpose
Next.js route segment for runtimes in the web client.

## Key Files
| File | Description |
| --- | --- |
| `page.tsx` | Source file local to this directory. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Keep route files thin: layout, route, and page modules should mostly compose existing view or content modules.
- Push reusable stateful logic down into shared packages or feature modules rather than growing route glue.

### Testing Requirements
- Run `pnpm typecheck` for TypeScript changes that affect shared types or cross-workspace imports.
- Run the closest Vitest target, or `pnpm test`, when changing UI behavior, hooks, or shared state.

### Common Patterns
- Route directories usually contain thin page, layout, or route handlers plus nested segments.
- Shared behavior should move into packages or helpers instead of growing the route shell.

## Dependencies

### Internal
- `packages/views/` and `packages/core/` hold most reusable product logic referenced by route shells.

### External
- `next` via `apps/web/package.json`.
- `react` via `apps/web/package.json`.
- `zustand` via `apps/web/package.json`.
- `@tanstack/react-query` via `apps/web/package.json`.
- `vitest` via `apps/web/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
