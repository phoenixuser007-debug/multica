<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# docs

## Purpose
Structured documentation pages grouped by audience and topic.

## Key Files
| File | Description |
| --- | --- |
| `index.mdx` | Directory-local file. |
| `meta.json` | Configuration file used by this area. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `cli/` | Documentation content for the cli topic area. See `cli/AGENTS.md`. |
| `developers/` | Documentation content for the developers topic area. See `developers/AGENTS.md`. |
| `getting-started/` | Documentation content for the getting started topic area. See `getting-started/AGENTS.md`. |
| `guides/` | Documentation content for the guides topic area. See `guides/AGENTS.md`. |

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
