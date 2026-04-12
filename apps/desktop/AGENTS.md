<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# desktop

## Purpose
Electron desktop application workspace for packaging Multica as a native desktop client.

## Key Files
| File | Description |
| --- | --- |
| `.gitignore` | Directory-local file. |
| `electron-builder.yml` | Electron Builder packaging configuration. |
| `electron.vite.config.ts` | Electron Vite configuration for desktop builds. |
| `eslint.config.mjs` | ESLint configuration for this workspace. |
| `package.json` | Package manifest with workspace name, scripts, and dependencies. |
| `tsconfig.json` | TypeScript configuration for this workspace. |
| `tsconfig.node.json` | TypeScript configuration for this workspace. |
| `tsconfig.web.json` | TypeScript configuration for this workspace. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `resources/` | Packaged desktop assets such as icons that are bundled with Electron builds. See `resources/AGENTS.md`. |
| `src/` | Desktop source code split by Electron process boundary. See `src/AGENTS.md`. |

## For AI Agents

### Working In This Directory
- Keep changes scoped to this directory and follow existing naming and layering patterns.
- Prefer extending existing modules over introducing parallel abstractions without a clear need.

### Testing Requirements
- Run the desktop workspace dev or build command after changing Electron process wiring.
- Verify changes across the correct Electron boundary: main, preload, or renderer.

### Common Patterns
- Follow the closest existing local pattern before inventing a new one.
- Prefer small, scoped edits so this directory stays easy to navigate.

## Dependencies

### Internal
- See the parent directory guidance in `../AGENTS.md` for adjacent modules that work with this area.

### External
- `react` via `apps/desktop/package.json`.
- `electron` via `apps/desktop/package.json`.
- `electron-vite` via `apps/desktop/package.json`.
- `typescript` via `apps/desktop/package.json`.
- `@dnd-kit/core` via `apps/desktop/package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
