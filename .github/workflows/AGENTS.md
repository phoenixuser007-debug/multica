<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# workflows

## Purpose
GitHub Actions workflows for CI, releases, and other repository automation.

## Key Files
| File | Description |
| --- | --- |
| `ci.yml` | Configuration file used by this area. |
| `release.yml` | Configuration file used by this area. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Treat workflow files as shared automation contracts; small syntax mistakes can break CI or releases.
- Keep contributor-facing templates aligned with the current developer workflow.

### Testing Requirements
- Spell-check and re-read rendered markdown after significant edits.
- Validate linked commands and paths against the current repository layout.

### Common Patterns
- Follow the closest existing local pattern before inventing a new one.
- Prefer small, scoped edits so this directory stays easy to navigate.

## Dependencies

### Internal
- See the parent directory guidance in `../AGENTS.md` for adjacent modules that work with this area.

### External
- `@playwright/test` via `package.json`.
- `typescript` via `package.json`.
- `@types/node` via `package.json`.
- `@types/pg` via `package.json`.
- `pg` via `package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
