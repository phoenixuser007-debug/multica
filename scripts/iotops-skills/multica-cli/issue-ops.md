# `multica issue` — full recipe book

All flags are discoverable via `multica issue <subcommand> --help`. Below are the agent-ready recipes with worked examples.

## Search for open duplicates (dedup)

```bash
# JSON output so subsequent jq steps can filter.
multica issue search "iotops-client-state-processor::NullPointerException" --output json \
  | jq 'map(select(.status != "done" and .status != "cancelled")) | .[0]'
```

- `--include-closed` if you want done/cancelled too (rarely useful for dedup).
- `--limit 20` default; raise for very-noisy fingerprints.

The query is **a substring match against title + description**. For strict fingerprinting, include the whole `cluster::service::exception::file:line` blob in the issue description so search hits exactly.

## List filtered

```bash
# Open issues assigned to a specific agent, newest first — for "has the Scout already run?" checks.
multica issue list --assignee "IoTOps Scout" --status todo --output json \
  | jq 'sort_by(.created_at) | reverse | .[0]'
```

Filters: `--assignee`, `--status`, `--priority`, `--project`. Paging via `--limit` / `--offset`.

## Get one issue

```bash
multica issue get "$ID" --output json | jq '{id, title, status, assignee_name, parent_issue_id}'
```

## Create a child issue (handoff)

```bash
# Scout → Fixer
multica issue create \
  --title "[$SERVICE] $JIRA_KEY: $EXC at $FILE:$LINE" \
  --description "$(cat /tmp/fixer-payload.md)" \
  --assignee "IoTOps Fixer" \
  --parent "$SCOUT_ISSUE_ID" \
  --priority high \
  --status todo \
  --output json

# Fixer → Verifier
multica issue create \
  --title "$JIRA_KEY: Raise PR for $SERVICE $EXC fix" \
  --description "$(cat /tmp/verifier-payload.md)" \
  --assignee "IoTOps Verifier" \
  --parent "$FIXER_ISSUE_ID" \
  --priority high \
  --status todo \
  --output json
```

**Why `--assignee` takes a name**: the CLI resolves the name to an agent UUID at request time, so the payload is portable — no hardcoded UUIDs, no stale references if the agent is re-provisioned.

Capture the returned `id` from stdout's JSON; you'll use it in the final comment on your own tick issue so the chain is navigable.

## Add a comment

```bash
# One-liner
multica issue comment add "$ID" --content "RCA confidence 0.9 — proceeding to TDD step."

# Multi-line (strongly prefer --content-stdin for anything with code blocks)
multica issue comment add "$ID" --content-stdin <<EOF
## Proposed Fix

**Root cause**: null-unsafe KTable lookup on an evicted reporter key.
**Approach**: guard with \`Optional.ofNullable\` + early return.
**Minimal change**: two lines in \`StateProcessor.java\`.
**Test strategy**: new JUnit test \`StateProcessorNullSafetyTest\` that feeds an evicted key and asserts no NPE.
EOF
```

## Update an issue (title / description / assignee / parent)

```bash
multica issue update "$ID" --status in_progress --output json
multica issue update "$ID" --title "New title" --description "Rewritten body"
multica issue update "$ID" --assignee "IoTOps Verifier"   # reassign
multica issue update "$ID" --parent "$OTHER_ISSUE_ID"     # change parent
multica issue update "$ID" --parent ""                     # clear parent
```

## Status shortcuts

```bash
multica issue status "$ID" in_progress
multica issue status "$ID" done
multica issue status "$ID" blocked       # triggers retry_on_blocked after 24h
```

Faster than `issue update --status ...` because it skips re-serializing the full issue.

## Assign / unassign

```bash
multica issue assign "$ID" --to "IoTOps Verifier"
multica issue assign "$ID" --unassign
```

## Run / execution history

```bash
multica issue runs "$ID"                    # list executions for this issue
multica issue run-messages "$RUN_ID"        # the agent's execution transcript
```

Useful for debugging "why did the Fixer give up here?" — the run-messages output shows every tool call the agent made.

## Idempotency patterns

The CLI doesn't retry automatically. Wrap mutating commands in:

```bash
retry() {
  local attempts="${1:-3}" sleep="${2:-2}"
  shift 2
  for i in $(seq 1 "$attempts"); do
    "$@" && return 0
    echo "retry $i/$attempts after $sleep s" >&2
    sleep "$sleep"
    sleep=$((sleep * 2))
  done
  return 1
}

retry 3 2 multica issue create --title "..." --description "..." \
  --assignee "IoTOps Fixer" --parent "$PARENT" --output json
```

For idempotent **reads** (search, list, get), plain retries are safe. For **create**, prefer search-then-create to avoid accidental duplicates during retries — the CLI doesn't yet accept an Idempotency-Key.
