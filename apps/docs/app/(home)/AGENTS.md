<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# (home)

## Purpose
Docs site route group for the public landing experience.

## Key Files
| File | Description |
| --- | --- |
| `layout.tsx` | Next.js layout component for this route segment. |
| `page.tsx` | Source file local to this directory. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Keep route files thin: layout, route, and page modules should mostly compose existing view or content modules.
- Push reusable stateful logic down into shared packages or feature modules rather than growing route glue.

### Testing Requirements
- Run the docs workspace build or dev command before shipping content or route changes.
- Verify affected pages and search behavior in the browser after changing docs routing or content loaders.

### Common Patterns
- Route directories usually contain thin page, layout, or route handlers plus nested segments.
- Shared behavior should move into packages or helpers instead of growing the route shell.

## Dependencies

### Internal
- `apps/docs/content/` and `apps/docs/lib/` provide the rendered docs source and loaders.

### External
- `next` via `apps/docs/package.json`.
- `react` via `apps/docs/package.json`.
- `typescript` via `apps/docs/package.json`.
- `@tailwindcss/postcss` via `apps/docs/package.json`.
- `@types/react` via `apps/docs/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
