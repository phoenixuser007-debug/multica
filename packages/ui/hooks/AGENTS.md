<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# hooks

## Purpose
Reusable hooks packaged with the shared UI layer.

## Key Files
| File | Description |
| --- | --- |
| `use-auto-scroll.ts` | Source file local to this directory. |
| `use-mobile.ts` | Source file local to this directory. |
| `use-scroll-fade.ts` | Source file local to this directory. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Keep hooks focused on one concern and isolate side effects from derived state.
- Expose a minimal API so calling components stay easy to test.

### Testing Requirements
- Run `pnpm typecheck` for TypeScript changes that affect shared types or cross-workspace imports.
- Run the closest Vitest target, or `pnpm test`, when changing UI behavior, hooks, or shared state.

### Common Patterns
- Separate derived state from side effects so hooks stay testable.
- Return a narrow surface area to keep call sites readable.

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
