<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# eslint-config

## Purpose
Reusable ESLint presets shared across the monorepo.

## Key Files
| File | Description |
| --- | --- |
| `base.js` | Source file local to this directory. |
| `next.js` | Source file local to this directory. |
| `package.json` | Package manifest with workspace name, scripts, and dependencies. |
| `react.js` | Source file local to this directory. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Keep changes scoped to this directory and follow existing naming and layering patterns.
- Prefer extending existing modules over introducing parallel abstractions without a clear need.

### Testing Requirements
- Run the closest relevant automated check for files changed in this directory.
- If there is no dedicated automation, validate imports, paths, and rendered output manually.

### Common Patterns
- Follow the closest existing local pattern before inventing a new one.
- Prefer small, scoped edits so this directory stays easy to navigate.

## Dependencies

### Internal
- See the parent directory guidance in `../AGENTS.md` for adjacent modules that work with this area.

### External
- `@eslint/js` via `packages/eslint-config/package.json`.
- `@next/eslint-plugin-next` via `packages/eslint-config/package.json`.
- `eslint-plugin-react` via `packages/eslint-config/package.json`.
- `eslint-plugin-react-hooks` via `packages/eslint-config/package.json`.
- `typescript-eslint` via `packages/eslint-config/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
