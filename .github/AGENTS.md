<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# .github

## Purpose
GitHub-specific repository metadata, automation hooks, and contributor-facing templates.

## Key Files
| File | Description |
| --- | --- |
| `PULL_REQUEST_TEMPLATE.md` | Repository documentation or design note. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `workflows/` | GitHub Actions workflows for CI, releases, and other repository automation. See `workflows/AGENTS.md`. |

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
- See the root `AGENTS.md` for top-level repository relationships.

### External
- `@playwright/test` via `package.json`.
- `typescript` via `package.json`.
- `@types/node` via `package.json`.
- `@types/pg` via `package.json`.
- `pg` via `package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
