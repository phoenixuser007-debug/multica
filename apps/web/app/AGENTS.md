<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# app

## Purpose
App Router entrypoints, layouts, and route-specific shells for the web client.

## Key Files
| File | Description |
| --- | --- |
| `custom.css` | Global styling entrypoint for this app surface. |
| `globals.css` | Global styling entrypoint for this app surface. |
| `layout.tsx` | Next.js layout component for this route segment. |
| `robots.ts` | Robots metadata for search engine crawling. |
| `sitemap.ts` | Sitemap generation for the web application. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `(auth)/` | Auth-related route group for the web client. See `(auth)/AGENTS.md`. |
| `(dashboard)/` | Authenticated dashboard route group for day-to-day product workflows. See `(dashboard)/AGENTS.md`. |
| `(landing)/` | Marketing and public-facing route group for the web client. See `(landing)/AGENTS.md`. |
| `auth/` | Auth callback and related direct auth routes. See `auth/AGENTS.md`. |
| `favicon.ico/` | Next.js route segment for favicon.ico in the web client. See `favicon.ico/AGENTS.md`. |

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
