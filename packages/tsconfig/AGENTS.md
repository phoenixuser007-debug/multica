<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# tsconfig

## Purpose
Reusable TypeScript base configs shared across workspaces.

## Key Files
| File | Description |
| --- | --- |
| `base.json` | Configuration file used by this area. |
| `package.json` | Package manifest with workspace name, scripts, and dependencies. |
| `react-library.json` | Configuration file used by this area. |

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
- Inherit external dependency context from the nearest parent workspace.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
