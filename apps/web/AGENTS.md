<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# web

## Purpose
Next.js web client workspace for the primary Multica browser experience.

## Key Files
| File | Description |
| --- | --- |
| `components.json` | shadcn component registry configuration. |
| `eslint.config.mjs` | ESLint configuration for this workspace. |
| `next-env.d.ts` | Source file local to this directory. |
| `next.config.ts` | Next.js build and runtime configuration. |
| `package.json` | Package manifest with workspace name, scripts, and dependencies. |
| `postcss.config.mjs` | PostCSS configuration for styling pipelines. |
| `proxy.ts` | Proxy or runtime forwarding configuration for the app. |
| `tsconfig.json` | TypeScript configuration for this workspace. |
| `vitest.config.ts` | Vitest configuration for local unit tests. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `app/` | App Router entrypoints, layouts, and route-specific shells for the web client. See `app/AGENTS.md`. |
| `components/` | Web-app-only composition components and provider wrappers. See `components/AGENTS.md`. |
| `features/` | App-local feature modules that have not been extracted into shared packages. See `features/AGENTS.md`. |
| `platform/` | Platform-specific adapters used by the web app shell. See `platform/AGENTS.md`. |
| `public/` | Static assets served directly by the web application. See `public/AGENTS.md`. |
| `test/` | Shared test setup and helpers for web-unit coverage. See `test/AGENTS.md`. |

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
- `next` via `apps/web/package.json`.
- `react` via `apps/web/package.json`.
- `zustand` via `apps/web/package.json`.
- `@tanstack/react-query` via `apps/web/package.json`.
- `vitest` via `apps/web/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
