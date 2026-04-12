<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# components

## Purpose
Components view module containing UI composition for that feature area.

## Key Files
| File | Description |
| --- | --- |
| `account-tab.tsx` | Source file local to this directory. |
| `appearance-tab.tsx` | Source file local to this directory. |
| `index.ts` | Barrel export for the directory or package. |
| `members-tab.tsx` | Source file local to this directory. |
| `repositories-tab.tsx` | Source file local to this directory. |
| `settings-page.tsx` | Source file local to this directory. |
| `tokens-tab.tsx` | Source file local to this directory. |
| `workspace-tab.tsx` | Source file local to this directory. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Prefer composition over one-off abstractions and keep props explicit.
- Reuse shared UI primitives before introducing new bespoke building blocks.

### Testing Requirements
- Run `pnpm typecheck` for TypeScript changes that affect shared types or cross-workspace imports.
- Run the closest Vitest target, or `pnpm test`, when changing UI behavior, hooks, or shared state.

### Common Patterns
- Co-locate small presentation helpers with the component that owns them.
- Keep props explicit and favor composition over inheritance-style abstractions.

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
