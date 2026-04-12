<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# docs

## Purpose
Repository-level documentation, product notes, and implementation plans outside the shipped docs site.

## Key Files
No direct source files live here; this directory mainly organizes subdirectories.

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `assets/` | Static design and marketing assets referenced by repository docs. See `assets/AGENTS.md`. |
| `plans/` | Date-stamped engineering plans and design notes. See `plans/AGENTS.md`. |

## For AI Agents

### Working In This Directory
- Keep notes concise, dated when appropriate, and grounded in the current repo state.
- Preserve source-faithful wording when it improves future retrieval or grepability.

### Testing Requirements
- Spell-check and re-read rendered markdown after significant edits.
- Validate linked commands and paths against the current repository layout.

### Common Patterns
- Use dated filenames for temporal plans or review artifacts.
- Preserve concrete commands, paths, and evidence when documenting operational work.

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
