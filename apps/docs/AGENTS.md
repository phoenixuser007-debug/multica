<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# docs

## Purpose
Next.js documentation site workspace for product and developer docs.

## Key Files
| File | Description |
| --- | --- |
| `.gitignore` | Directory-local file. |
| `next-env.d.ts` | Source file local to this directory. |
| `next.config.mjs` | Next.js build and runtime configuration. |
| `package.json` | Package manifest with workspace name, scripts, and dependencies. |
| `postcss.config.mjs` | PostCSS configuration for styling pipelines. |
| `source.config.ts` | Docs content-source configuration. |
| `tsconfig.json` | TypeScript configuration for this workspace. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `app/` | Docs site routes, layouts, and API endpoints. See `app/AGENTS.md`. |
| `content/` | Source content consumed by the docs site. See `content/AGENTS.md`. |
| `lib/` | Helpers for loading, indexing, and shaping documentation content. See `lib/AGENTS.md`. |

## For AI Agents

### Working In This Directory
- Keep changes scoped to this directory and follow existing naming and layering patterns.
- Prefer extending existing modules over introducing parallel abstractions without a clear need.

### Testing Requirements
- Run the docs workspace build or dev command before shipping content or route changes.
- Verify affected pages and search behavior in the browser after changing docs routing or content loaders.

### Common Patterns
- Follow the closest existing local pattern before inventing a new one.
- Prefer small, scoped edits so this directory stays easy to navigate.

## Dependencies

### Internal
- See the parent directory guidance in `../AGENTS.md` for adjacent modules that work with this area.

### External
- `next` via `apps/docs/package.json`.
- `react` via `apps/docs/package.json`.
- `typescript` via `apps/docs/package.json`.
- `@tailwindcss/postcss` via `apps/docs/package.json`.
- `@types/react` via `apps/docs/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
