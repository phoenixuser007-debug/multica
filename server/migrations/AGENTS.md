<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# migrations

## Purpose
Ordered PostgreSQL schema migrations applied by the backend.

## Key Files
| File | Description |
| --- | --- |
| `00x_*.up.sql` | Forward PostgreSQL migrations applied in order. |
| `00x_*.down.sql` | Rollback partners for the forward migrations. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Add forward and rollback migration pairs with monotonic numbering.
- Keep migrations narrowly scoped so schema changes are easy to reason about and revert.

### Testing Requirements
- Run `cd server && go test ./...` or a targeted `go test` package for backend changes.
- Regenerate sqlc output with `make sqlc` after editing files under `server/pkg/db/queries/`.

### Common Patterns
- Follow the closest existing local pattern before inventing a new one.
- Prefer small, scoped edits so this directory stays easy to navigate.

## Dependencies

### Internal
- See the parent directory guidance in `../AGENTS.md` for adjacent modules that work with this area.

### External
- `github.com/go-chi/chi` for HTTP routing.
- `github.com/gorilla/websocket` for realtime transport.
- `sqlc`-generated database access driven by the SQL sources in this repo.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
