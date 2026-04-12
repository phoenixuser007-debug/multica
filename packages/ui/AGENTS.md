<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# ui

## Purpose
Shared UI primitives, styles, and markdown helpers.

## Key Files
| File | Description |
| --- | --- |
| `components.json` | shadcn component registry configuration. |
| `eslint.config.mjs` | ESLint configuration for this workspace. |
| `package.json` | Package manifest with workspace name, scripts, and dependencies. |
| `tsconfig.json` | TypeScript configuration for this workspace. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `components/` | Shared component exports organized by UI concern. See `components/AGENTS.md`. |
| `hooks/` | Reusable hooks packaged with the shared UI layer. See `hooks/AGENTS.md`. |
| `lib/` | Utility helpers supporting the shared UI package. See `lib/AGENTS.md`. |
| `markdown/` | Markdown rendering helpers and related UI integration. See `markdown/AGENTS.md`. |
| `styles/` | Shared style tokens and CSS entrypoints. See `styles/AGENTS.md`. |

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
