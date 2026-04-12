<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# editor

## Purpose
Editor view module containing UI composition for that feature area.

## Key Files
| File | Description |
| --- | --- |
| `content-editor.css` | Directory-local file. |
| `content-editor.tsx` | Source file local to this directory. |
| `file-drop-overlay.tsx` | Source file local to this directory. |
| `index.ts` | Barrel export for the directory or package. |
| `readonly-content.tsx` | Source file local to this directory. |
| `title-editor.css` | Directory-local file. |
| `title-editor.tsx` | Source file local to this directory. |
| `use-file-drop-zone.ts` | Source file local to this directory. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `extensions/` | Extensions view module containing UI composition for that feature area. See `extensions/AGENTS.md`. |
| `utils/` | Utils view module containing UI composition for that feature area. See `utils/AGENTS.md`. |

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
- `packages/core/` provides domain state and API access used by views.
- `packages/ui/` provides shared presentational primitives.

### External
- `vitest` via `packages/views/package.json`.
- `typescript` via `packages/views/package.json`.
- `@base-ui/react` via `packages/views/package.json`.
- `@dnd-kit/core` via `packages/views/package.json`.
- `@dnd-kit/sortable` via `packages/views/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
