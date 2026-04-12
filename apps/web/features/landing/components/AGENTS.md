<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# components

## Purpose
UI components used by the landing feature.

## Key Files
| File | Description |
| --- | --- |
| `about-page-client.tsx` | Source file local to this directory. |
| `changelog-page-client.tsx` | Source file local to this directory. |
| `faq-section.tsx` | Source file local to this directory. |
| `features-section.tsx` | Source file local to this directory. |
| `how-it-works-section.tsx` | Source file local to this directory. |
| `landing-footer.tsx` | Source file local to this directory. |
| `landing-header.tsx` | Source file local to this directory. |
| `landing-hero.tsx` | Source file local to this directory. |
| `multica-landing.tsx` | Source file local to this directory. |
| `open-source-section.tsx` | Source file local to this directory. |
| `shared.tsx` | Source file local to this directory. |

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
- `apps/web/components/` and `packages/ui/` provide shared UI composition primitives.

### External
- `next` via `apps/web/package.json`.
- `react` via `apps/web/package.json`.
- `zustand` via `apps/web/package.json`.
- `@tanstack/react-query` via `apps/web/package.json`.
- `vitest` via `apps/web/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
