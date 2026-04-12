<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# common

## Purpose
Common higher-level shared components.

## Key Files
| File | Description |
| --- | --- |
| `actor-avatar.tsx` | Source file local to this directory. |
| `emoji-picker.tsx` | Source file local to this directory. |
| `file-upload-button.tsx` | Source file local to this directory. |
| `mention-hover-card.tsx` | Source file local to this directory. |
| `multica-icon.tsx` | Source file local to this directory. |
| `quick-emoji-picker.tsx` | Source file local to this directory. |
| `reaction-bar.tsx` | Source file local to this directory. |
| `theme-provider.tsx` | Source file local to this directory. |

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
- `typescript` via `packages/ui/package.json`.
- `@base-ui/react` via `packages/ui/package.json`.
- `@emoji-mart/data` via `packages/ui/package.json`.
- `@multica/tsconfig` via `packages/ui/package.json`.
- `@types/linkify-it` via `packages/ui/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
