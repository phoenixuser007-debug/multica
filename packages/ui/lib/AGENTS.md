<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# lib

## Purpose
Utility helpers supporting the shared UI package.

## Key Files
| File | Description |
| --- | --- |
| `utils.ts` | Source file local to this directory. |

## Subdirectories
No documented subdirectories.

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
- See the parent directory guidance in `../AGENTS.md` for adjacent modules that work with this area.

### External
- `typescript` via `packages/ui/package.json`.
- `@base-ui/react` via `packages/ui/package.json`.
- `@emoji-mart/data` via `packages/ui/package.json`.
- `@multica/tsconfig` via `packages/ui/package.json`.
- `@types/linkify-it` via `packages/ui/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
