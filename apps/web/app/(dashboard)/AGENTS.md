<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# (dashboard)

## Purpose
Authenticated dashboard route group for day-to-day product workflows.

## Key Files
| File | Description |
| --- | --- |
| `layout.tsx` | Next.js layout component for this route segment. |
| `loading.tsx` | Source file local to this directory. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `agents/` | Next.js route segment for agents in the web client. See `agents/AGENTS.md`. |
| `inbox/` | Next.js route segment for inbox in the web client. See `inbox/AGENTS.md`. |
| `issues/` | Next.js route segment for issues in the web client. See `issues/AGENTS.md`. |
| `my-issues/` | Next.js route segment for my issues in the web client. See `my-issues/AGENTS.md`. |
| `projects/` | Next.js route segment for projects in the web client. See `projects/AGENTS.md`. |
| `runtimes/` | Next.js route segment for runtimes in the web client. See `runtimes/AGENTS.md`. |
| `settings/` | Next.js route segment for settings in the web client. See `settings/AGENTS.md`. |
| `skills/` | Next.js route segment for skills in the web client. See `skills/AGENTS.md`. |

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
