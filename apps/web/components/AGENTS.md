<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# components

## Purpose
Web-app-only composition components and provider wrappers.

## Key Files
| File | Description |
| --- | --- |
| `locale-sync.tsx` | Source file local to this directory. |
| `theme-provider.tsx` | Source file local to this directory. |
| `web-providers.tsx` | Source file local to this directory. |

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
- See the parent directory guidance in `../AGENTS.md` for adjacent modules that work with this area.

### External
- `next` via `apps/web/package.json`.
- `react` via `apps/web/package.json`.
- `zustand` via `apps/web/package.json`.
- `@tanstack/react-query` via `apps/web/package.json`.
- `vitest` via `apps/web/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
