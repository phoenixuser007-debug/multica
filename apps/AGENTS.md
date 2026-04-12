<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# apps

## Purpose
Application workspaces in the monorepo, including the browser client, docs site, and desktop shell.

## Key Files
No direct source files live here; this directory mainly organizes subdirectories.

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `desktop/` | Electron desktop application workspace for packaging Multica as a native desktop client. See `desktop/AGENTS.md`. |
| `docs/` | Next.js documentation site workspace for product and developer docs. See `docs/AGENTS.md`. |
| `web/` | Next.js web client workspace for the primary Multica browser experience. See `web/AGENTS.md`. |

## For AI Agents

### Working In This Directory
- Keep changes scoped to this directory and follow existing naming and layering patterns.
- Prefer extending existing modules over introducing parallel abstractions without a clear need.

### Testing Requirements
- Run the closest relevant automated check for files changed in this directory.
- If there is no dedicated automation, validate imports, paths, and rendered output manually.

### Common Patterns
- Follow the closest existing local pattern before inventing a new one.
- Prefer small, scoped edits so this directory stays easy to navigate.

## Dependencies

### Internal
- See the root `AGENTS.md` for top-level repository relationships.

### External
- `@playwright/test` via `package.json`.
- `typescript` via `package.json`.
- `@types/node` via `package.json`.
- `@types/pg` via `package.json`.
- `pg` via `package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
