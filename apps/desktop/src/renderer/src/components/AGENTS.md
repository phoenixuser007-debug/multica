<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# components

## Purpose
Desktop renderer UI components.

## Key Files
| File | Description |
| --- | --- |
| `desktop-layout.tsx` | Source file local to this directory. |
| `tab-bar.tsx` | Source file local to this directory. |
| `tab-content.tsx` | Source file local to this directory. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Prefer composition over one-off abstractions and keep props explicit.
- Reuse shared UI primitives before introducing new bespoke building blocks.

### Testing Requirements
- Run the desktop workspace dev or build command after changing Electron process wiring.
- Verify changes across the correct Electron boundary: main, preload, or renderer.

### Common Patterns
- Co-locate small presentation helpers with the component that owns them.
- Keep props explicit and favor composition over inheritance-style abstractions.

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
