<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# src

## Purpose
React renderer source for the desktop client UI.

## Key Files
| File | Description |
| --- | --- |
| `App.tsx` | Source file local to this directory. |
| `env.d.ts` | Source file local to this directory. |
| `globals.css` | Global styling entrypoint for this app surface. |
| `main.tsx` | Source file local to this directory. |
| `routes.tsx` | Source file local to this directory. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `components/` | Desktop renderer UI components. See `components/AGENTS.md`. |
| `hooks/` | Desktop renderer React hooks. See `hooks/AGENTS.md`. |
| `pages/` | Top-level desktop renderer pages and route screens. See `pages/AGENTS.md`. |
| `platform/` | Platform integration helpers for the desktop renderer. See `platform/AGENTS.md`. |
| `stores/` | Desktop renderer state stores and client state helpers. See `stores/AGENTS.md`. |

## For AI Agents

### Working In This Directory
- Keep changes scoped to this directory and follow existing naming and layering patterns.
- Prefer extending existing modules over introducing parallel abstractions without a clear need.

### Testing Requirements
- Run the desktop workspace dev or build command after changing Electron process wiring.
- Verify changes across the correct Electron boundary: main, preload, or renderer.

### Common Patterns
- Follow the closest existing local pattern before inventing a new one.
- Prefer small, scoped edits so this directory stays easy to navigate.

## Dependencies

### Internal
- See the parent directory guidance in `../AGENTS.md` for adjacent modules that work with this area.

### External
- `react` via `apps/desktop/package.json`.
- `electron` via `apps/desktop/package.json`.
- `electron-vite` via `apps/desktop/package.json`.
- `typescript` via `apps/desktop/package.json`.
- `@dnd-kit/core` via `apps/desktop/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
