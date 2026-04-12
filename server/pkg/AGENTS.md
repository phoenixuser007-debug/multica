<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# pkg

## Purpose
Reusable backend packages and generated database access layers.

## Key Files
No direct source files live here; this directory mainly organizes subdirectories.

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `agent/` | Agent runtime abstraction and provider integrations. See `agent/AGENTS.md`. |
| `db/` | Database-facing packages including generated sqlc code and query sources. See `db/AGENTS.md`. |
| `protocol/` | Shared protocol definitions for backend integrations. See `protocol/AGENTS.md`. |
| `redact/` | Redaction helpers for logs or serialized backend output. See `redact/AGENTS.md`. |

## For AI Agents

### Working In This Directory
- Keep changes scoped to this directory and follow existing naming and layering patterns.
- Prefer extending existing modules over introducing parallel abstractions without a clear need.

### Testing Requirements
- Run `cd server && go test ./...` or a targeted `go test` package for backend changes.
- Regenerate sqlc output with `make sqlc` after editing files under `server/pkg/db/queries/`.

### Common Patterns
- Keep business logic testable and push IO to the edges when possible.
- Prefer small domain-oriented files over catch-all utility packages.

## Dependencies

### Internal
- See the parent directory guidance in `../AGENTS.md` for adjacent modules that work with this area.

### External
- `github.com/go-chi/chi` for HTTP routing.
- `github.com/gorilla/websocket` for realtime transport.
- `sqlc`-generated database access driven by the SQL sources in this repo.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
