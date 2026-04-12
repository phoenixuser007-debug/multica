<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# generated

## Purpose
Generated sqlc output. Do not hand-edit files here.

## Key Files
| File | Description |
| --- | --- |
| `*.go` | 25 files with the .go extension. |

## Subdirectories
No documented subdirectories.

## For AI Agents

### Working In This Directory
- Do not hand-edit generated sqlc output here.
- Change SQL under `server/pkg/db/queries/` or migrations, then regenerate artifacts instead.

### Testing Requirements
- Run `cd server && go test ./...` or a targeted `go test` package for backend changes.
- Regenerate sqlc output with `make sqlc` after editing files under `server/pkg/db/queries/`.

### Common Patterns
- Keep business logic testable and push IO to the edges when possible.
- Prefer small domain-oriented files over catch-all utility packages.

## Dependencies

### Internal
- Generated from `server/pkg/db/queries/` using `server/sqlc.yaml`.

### External
- `github.com/go-chi/chi` for HTTP routing.
- `github.com/gorilla/websocket` for realtime transport.
- `sqlc`-generated database access driven by the SQL sources in this repo.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
