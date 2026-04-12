<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# src

## Purpose
Desktop source code split by Electron process boundary.

## Key Files
No direct source files live here; this directory mainly organizes subdirectories.

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `main/` | Electron main-process startup, window management, and native integration code. See `main/AGENTS.md`. |
| `preload/` | Secure preload bridge exposed from Electron main to renderer. See `preload/AGENTS.md`. |
| `renderer/` | Renderer-process workspace used by the desktop shell. See `renderer/AGENTS.md`. |

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
