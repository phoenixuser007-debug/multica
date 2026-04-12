<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# components

## Purpose
Shared component exports organized by UI concern.

## Key Files
No direct source files live here; this directory mainly organizes subdirectories.

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `common/` | Common higher-level shared components. See `common/AGENTS.md`. |
| `ui/` | Low-level shadcn-style UI primitives. See `ui/AGENTS.md`. |

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
- See the parent directory guidance in `../AGENTS.md` for adjacent modules that work with this area.

### External
- `typescript` via `packages/ui/package.json`.
- `@base-ui/react` via `packages/ui/package.json`.
- `@emoji-mart/data` via `packages/ui/package.json`.
- `@multica/tsconfig` via `packages/ui/package.json`.
- `@types/linkify-it` via `packages/ui/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
