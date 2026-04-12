<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# internal

## Purpose
Internal backend packages implementing domain logic, handlers, middleware, and runtime orchestration.

## Key Files
No direct source files live here; this directory mainly organizes subdirectories.

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `auth/` | Authentication helpers and token lifecycle code. See `auth/AGENTS.md`. |
| `cli/` | Internal backend package for cli concerns. See `cli/AGENTS.md`. |
| `daemon/` | Local runtime loop that discovers runtimes, polls for tasks, and manages agent execution. See `daemon/AGENTS.md`. |
| `events/` | Internal event bus primitives for decoupled backend communication. See `events/AGENTS.md`. |
| `handler/` | HTTP handlers for product domains and daemon endpoints. See `handler/AGENTS.md`. |
| `logger/` | Structured logging helpers and setup. See `logger/AGENTS.md`. |
| `mention/` | Internal backend package for mention concerns. See `mention/AGENTS.md`. |
| `middleware/` | HTTP middleware for auth, workspace selection, and request shaping. See `middleware/AGENTS.md`. |
| `realtime/` | WebSocket hub and realtime broadcast infrastructure. See `realtime/AGENTS.md`. |
| `sanitize/` | Internal backend package for sanitize concerns. See `sanitize/AGENTS.md`. |
| `service/` | Backend service-layer orchestration for tasks and related workflows. See `service/AGENTS.md`. |
| `storage/` | Storage-related backend helpers. See `storage/AGENTS.md`. |
| `util/` | Internal backend package for util concerns. See `util/AGENTS.md`. |

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
