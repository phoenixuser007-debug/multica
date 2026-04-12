<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# hooks

## Purpose
Desktop renderer React hooks.

## Key Files
| File | Description |
| --- | --- |
| `use-document-title.ts` | Source file local to this directory. |
| `use-tab-history.ts` | Source file local to this directory. |
| `use-tab-router-sync.ts` | Source file local to this directory. |
| `use-tab-sync.ts` | Source file local to this directory. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Keep hooks focused on one concern and isolate side effects from derived state.
- Expose a minimal API so calling components stay easy to test.

### Testing Requirements
- Run the desktop workspace dev or build command after changing Electron process wiring.
- Verify changes across the correct Electron boundary: main, preload, or renderer.

### Common Patterns
- Separate derived state from side effects so hooks stay testable.
- Return a narrow surface area to keep call sites readable.

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
