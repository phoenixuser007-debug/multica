<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-11 | Updated: 2026-04-11 -->

# daemon

## Purpose
Local runtime loop that discovers runtimes, polls for tasks, and manages agent execution.

## Key Files
| File | Description |
| --- | --- |
| `client.go` | Directory-local file. |
| `config.go` | Directory-local file. |
| `daemon_test.go` | Directory-local file. |
| `daemon.go` | Directory-local file. |
| `health.go` | Directory-local file. |
| `helpers.go` | Directory-local file. |
| `prompt.go` | Directory-local file. |
| `types.go` | Directory-local file. |

## Subdirectories
| Directory | Purpose |
| --- | --- |
| `execenv/` | Execution-environment helpers used by the daemon when spawning agent work. See `execenv/AGENTS.md`. |
| `repocache/` | Repository cache helpers used by daemon task execution. See `repocache/AGENTS.md`. |
| `usage/` | Runtime usage accounting and reporting helpers for daemon execution. See `usage/AGENTS.md`. |

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
- `server/pkg/agent/` provides runtime backends used by daemon execution.

### External
- `github.com/go-chi/chi` for HTTP routing.
- `github.com/gorilla/websocket` for realtime transport.
- `sqlc`-generated database access driven by the SQL sources in this repo.

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
