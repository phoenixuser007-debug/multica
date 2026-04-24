---
name: multica-cli
description: "Drive the multica workspace (issues, agents, skills, comments, status transitions) from the command line via the `multica` binary. Use whenever you need to create a child issue for another agent, dedup via search, add an audit comment, transition your own issue status, or look up an agent/skill by name. Canonical alternative to raw REST curl — handles auth + workspace scoping automatically."
allowed-tools: ["Bash", "Read"]
---

# multica-cli

The `multica` CLI is the canonical way for agents to read and mutate workspace state. It's already installed on the runtime host and authenticated via `/root/.multica/config.json` — no tokens to pass, no workspace IDs to thread through, no URL-encoding traps.

**Use this skill instead of curling the REST API directly** — the CLI handles auth, workspace scoping, and encoding automatically, and error messages are much clearer.

## When to use

Any time you need to interact with the multica workspace from an agent:

| Need | Command |
|---|---|
| Check if a similar issue already exists (dedup) | `multica issue search "<fingerprint>"` |
| Create a child issue for another agent | `multica issue create --title ... --description ... --assignee "<agent name>" --parent <parent-id>` |
| Add an audit/progress comment on the current issue | `multica issue comment add <issue-id> --content "..."` |
| Transition your own issue to `done` / `blocked` / `in_progress` | `multica issue status <issue-id> <new-status>` |
| Look up an agent by name (no hardcoding UUIDs) | `multica agent list --output json \| jq '.[] \| select(.name=="IoTOps Fixer") \| .id'` |
| Inspect a skill attached to your agent | `multica skill get <id>` |

## Auth

The CLI reads `/root/.multica/config.json`:

```json
{
  "server_url": "http://localhost:8090",
  "workspace_id": "4b3bee90-…",
  "token": "mul_…"
}
```

You don't need to pass `--token`, `--workspace-id`, or `--server-url` — they're picked up automatically. If the file is missing or the token is invalid, the CLI emits a clear error (no need to parse HTTP status codes).

## Core recipes

See `issue-ops.md` for the full recipe set with worked examples. The most common three:

### 1. Spawn a child issue for another agent (Scout → Fixer, Fixer → Verifier)

```bash
multica issue create \
  --title "[$SERVICE] $JIRA_KEY: $EXC at $FILE:$LINE" \
  --description "$(cat /tmp/fixer-payload.md)" \
  --assignee "IoTOps Fixer" \
  --parent "$CURRENT_ISSUE_ID" \
  --priority high --status todo \
  --output json
```

Capture the returned JSON's `id` field — that's the new Fixer issue. Pass it back up in your own finishing comment so the chain is browsable.

### 2. Dedupe via search before creating

```bash
# Open issues (status != done/cancelled) matching a fingerprint string.
multica issue search "${SERVICE}::${EXCEPTION}::${FILE}:${LINE}" --output json \
  | jq 'map(select(.status != "done" and .status != "cancelled"))'
```

If the filtered array is non-empty, an existing chain is already on this error — skip creation and add a skip comment to the current tick issue instead.

### 3. Add an audit comment

```bash
multica issue comment add "$CURRENT_ISSUE_ID" --content-stdin <<EOF
## RCA

Root cause: NullPointerException on reporter KTable lookup after TTL eviction.
Confidence: 0.9
Affected file: src/main/java/com/aruba/iotops/StateProcessor.java:187
EOF
```

Prefer `--content-stdin` over `--content` when the body has backticks, double quotes, or multiple lines — avoids shell-escaping bugs.

## Status transitions

```bash
multica issue status "$ID" todo          # unstarted
multica issue status "$ID" in_progress   # claimed
multica issue status "$ID" in_review     # handed off (e.g. Verifier opened the PR)
multica issue status "$ID" blocked       # needs human — retry_on_blocked re-enqueues in 24h
multica issue status "$ID" done          # terminal success
multica issue status "$ID" cancelled     # terminal without success; no retry
```

Pick `blocked` (not `done`) when something needs human attention — the autopilot scheduler's `retry_on_blocked` only fires for `blocked`. Use `done` for terminal success, `cancelled` for "this wasn't actionable."

## Output format

Always pass `--output json` when the output is machine-consumed (your own subsequent step), not when you're displaying to a human. The default for `list` / `search` / `status` is `table` (pretty but not parseable); `get` / `create` / `update` default to `json`.

## Things not to do

- Do **not** call `curl -H "Authorization: ..." $MULTICA_API_URL/api/...` when a CLI subcommand exists. The CLI is maintained alongside the server; REST shape can drift from your hardcoded URL over versions.
- Do **not** hardcode agent UUIDs. Use `--assignee "<name>"` on `issue create` / `issue assign` — the CLI resolves the name to an ID.
- Do **not** use `multica` inside the `dev-env-dev-1` container — the container doesn't have the daemon config. Run `multica` from the host side of any `docker exec` boundary.
- Do **not** silently swallow CLI errors. If a `multica issue create` fails, log `stderr` and fall back to posting a comment on the parent tick issue so the human can see what went wrong.
