<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# scripts

## Purpose
Directory for scripts within the repository structure.

## Key Files
| File | Description |
| --- | --- |
| `check.sh` | Directory-local file. |
| `dev.sh` | Directory-local file. |
| `ensure-postgres.sh` | Directory-local file. |
| `init-worktree-env.sh` | Directory-local file. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Keep changes scoped to this directory and follow existing naming and layering patterns.
- Prefer extending existing modules over introducing parallel abstractions without a clear need.

### Testing Requirements
- Run the modified script directly in a safe local context before relying on it in automation.
- Prefer idempotent checks because these files are typically used by setup and CI flows.

### Common Patterns
- Follow the closest existing local pattern before inventing a new one.
- Prefer small, scoped edits so this directory stays easy to navigate.

## Dependencies

### Internal
- Used by the root `Makefile`, local setup flow, and CI automation.

### External
- `@playwright/test` via `package.json`.
- `typescript` via `package.json`.
- `@types/node` via `package.json`.
- `@types/pg` via `package.json`.
- `pg` via `package.json`.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
