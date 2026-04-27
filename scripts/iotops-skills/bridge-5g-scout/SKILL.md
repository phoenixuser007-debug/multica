---
name: bridge-5g-scout
description: "Daily scan of 4 Humio clusters (aqua, brooke, jedi, firth) for bridge-5g-* runtime errors. For each distinct error: fetch ±10-min pod context, look up the affected source file(s) in Stash, classify by fingerprint, create a JIRA CNX bug (or reuse an existing one), and hand off to the IoTOps Fixer agent as a child multica issue. Triggers on issues whose title starts with 'Scan bridge-5g-*' — created by the Bridge 5G Error Scout autopilot."
allowed-tools: ["Read", "Bash", "Grep", "Glob"]
---

# Bridge 5G Scout

You are the **first link** in the 3-agent Bridge 5G error chain. An autopilot fires you once a day (schedule set in the multica UI). Your job: scan 4 clusters, classify each distinct error by looking at the actual code, create Jira + Fixer child issue, hand off.

## When this skill fires

Your issue title starts with `Scan bridge-5g-*`. The issue was created by the **Bridge 5G Error Scout autopilot** on its scheduled trigger.

## Prerequisites (env on the runtime host)

Per-cluster Humio Bearer tokens:
- `AQUA_HUMIO_TOKEN`, `BROOKE_HUMIO_TOKEN`, `JEDI_HUMIO_TOKEN`, `FIRTH_HUMIO_TOKEN`

Optional URL overrides (only when the default cloudops pattern doesn't apply):
- `JEDI_HUMIO_URL`, `FIRTH_HUMIO_URL` (the script falls back to `https://<cluster>.cloudops.arubadev.cloud.hpe.com/logs` otherwise).

Other:
- `BITBUCKET_STASH_TOKEN` — Stash PAT (fetches affected source files for classification)
- `JIRA_TOKEN`, `JIRA_USERNAME` — for the Jira Bug creation step
- `MULTICA_API_URL`, `MULTICA_APP_URL`, `MULTICA_TOKEN`, `MULTICA_WORKSPACE_ID`, `IOTOPS_FIXER_AGENT_ID` — to create the child issue
- `SLACK_BOT_TOKEN` + `SLACK_CHANNEL_ID` — Bot API token and channel for threaded notifications (optional; falls back to `SLACK_WEBHOOK_URL` without threading if unset)
- `SLACK_WEBHOOK_URL` — fallback webhook if bot token is not configured
- `python3` with `requests` on PATH
- The bundled helper script `scripts/humio_client.py` (materialized to `.claude/skills/humio-skill/scripts/humio_client.py` at task-claim time — attached via the `humio-skill` you also have).

## Steps

### 1. Determine the window

The autopilot runs daily. Use a **24-hour sliding window ending at `now`** by default.

To find the start of the window: search for the previous Scout tick via `multica issue search "Scan bridge-5g-*" --output json`, filter to issues with `status == "done"` or `status == "in_review"` (i.e. **not** `blocked` or `cancelled`), sort by `created_at` descending, and use the most recent successful tick's `created_at` as `start`. If no successful previous tick exists, fall back to 24 h ago.

**Important:** only advance the window from a tick that successfully completed (reached `done`/`in_review`). A `blocked` or `cancelled` tick did not create Fixer issues, so its timestamp must **not** be used as the new window start — doing so would silently drop all the errors that tick found.

### 2. Scan each of 4 clusters

Iterate `for cluster in aqua brooke jedi firth`:

```bash
python3 $HOME/.claude/skills/humio-skill/scripts/humio_client.py \
  -c "$cluster" --iotops-scan --start "$START_ISO" --end now \
  -o "/tmp/errors-$cluster.json"
```

Each run emits a JSON array of fingerprints:

```json
{"pod": "...", "service_repo": "bridge-5g-state-processor",
 "exception_class": "NullPointerException",
 "timestamp_ms": 1713950400000,
 "raw_log": "...", "event_id": "..."}
```

If a cluster's token isn't set, log a warning and skip that cluster — do NOT fail the whole run. Other clusters should still process.

### 3. Merge + tag each event with its source cluster

Concatenate `/tmp/errors-*.json` into one list, annotating each entry with its `cluster` (aqua / brooke / jedi / firth) so later steps can fetch the right Humio context.

### 4. For each error event — fetch ±10-min pod context

```bash
python3 $HOME/.claude/skills/humio-skill/scripts/humio_client.py \
  -c "$cluster" --pod-context --pod "$POD" --timestamp-ms "$TS_MS" \
  --window-minutes 10 -o "/tmp/context-$EVENT_ID.json"
```

### 5. Classify by looking up the code

The error log usually contains a Java stack frame like
```
at com.aruba.iotops.state.StateProcessor.processRecord(StateProcessor.java:187)
```
Extract the top application frame (first frame under `com.aruba.iotops.*`) → `file = StateProcessor.java`, `line = 187`.

Fetch that file from Stash (branch `devel`) to confirm it exists and capture the surrounding ±20 lines:

```bash
curl -sf -H "Authorization: Bearer $BITBUCKET_STASH_TOKEN" \
  "https://stash.arubanetworks.com/rest/api/1.0/projects/GVT/repos/$SERVICE_REPO/raw/$RELATIVE_PATH?at=refs/heads/devel"
```

If the file exists but the line is in a removed block (e.g. after a refactor), still pass the fetch — the stack frame may be slightly stale. Only give up if the file is 404.

### 6. Build fingerprint + dedupe

Fingerprint: `{cluster}::{service_repo}::{ExceptionClass}::{source_file}::{line_no}`.

Two events with different source files are **different errors**, even if same exception class. Two events with the same fingerprint are the same error.

Dedup via the `multica-cli` skill (see its `issue-ops.md`):

```bash
multica issue search "${FINGERPRINT}" --output json \
  | jq 'map(select(.status != "done" and .status != "cancelled")) | .[0]'
```

If the filtered result is non-empty, another run is already on it — skip.

### 7. For each distinct fingerprint — ensure a JIRA Bug exists

Look for an existing JIRA ticket first (JQL search):
```
project = CNX AND component = "Unified Client - Non IP"
  AND summary ~ "\"<service_repo>\" \"<ExceptionClass>\" \"<source_file>\""
  AND status != Closed AND status != Resolved
```

If found → reuse its key. Otherwise create a new Bug:
```
POST https://jira.arubanetworks.com/rest/api/2/issue
Authorization: Basic <base64 JIRA_USERNAME:JIRA_TOKEN>

fields:
  project:      {key: CNX}
  issuetype:    {name: Bug}
  summary:      "[<service_repo>] <ExceptionClass> at <source_file>:<line>"
  description:  (Jira wiki markup — error log + cluster + Humio deep-link + code snippet)
  priority:     {name: Major}
  components:   [{name: Unified Client - Non IP}]
  versions:     [{name: CNX-APR-2026}]
  fixVersions:  [{name: CNX-APR-2026}]
  labels:       [auto-detected, iotops-error-monitor]
```

Full payload template and custom-field discovery is in `jira-fields.md` (also bundled with `iotops-verify-pr-slack`, but the shape is identical).

Save the returned `key` → `$JIRA_KEY`.

### 8. Create the Fixer child issue

Use the `multica-cli` skill — `multica issue create` resolves the assignee by name and fills in workspace context automatically:

```bash
multica issue create \
  --title "[$SERVICE] $JIRA_KEY: $EXC at $FILE:$LINE" \
  --description "$(cat /tmp/fixer-payload.md)" \
  --assignee "IoTOps Fixer" \
  --parent "$SCOUT_ISSUE_ID" \
  --priority high --status todo \
  --output json
```

The description content must follow this template (the Fixer skill parses these exact headings):

```markdown
## Error Context

**Cluster**: `<aqua|brooke|jedi|firth>`
**Service**: `<service_repo>`
**Pod**: `<pod_name>`
**Exception**: `<ExceptionClass>`
**Source**: `<affected_repo>/<relative_path>:<line>`
**Humio**: <deep-link>
**Jira**: https://jira.arubanetworks.com/browse/<JIRA_KEY>

## Error Log
```
<raw error log, max 3000 chars>
```

## ±10-min Pod Context
```
<from /tmp/context-*.json, flattened, max 200 lines>
```

## Affected Source (±20 lines around the stack frame)
```java
<from Stash fetch>
```

## Architecture

<contents of ARCHITECTURE.md from Stash, or "_not available_">
```

Title: `[<service_repo>] <JIRA_KEY>: <ExceptionClass> at <source_file>:<line>`.

### 8a. Post initial Slack message and propagate thread_ts

Immediately after the Fixer child issue is created, post a Slack message so the
team knows an error was detected. If `SLACK_BOT_TOKEN` + `SLACK_CHANNEL_ID` are
set the `slack_post` function (from `iotops-verify-pr-slack/slack-templates.md`)
returns a `ts` — store it in a multica comment on the Fixer issue so the Fixer
and Verifier can thread their replies.

```bash
# Load slack_post and extract_thread_ts from the shared template file
# (the skill runtime materialises all attached skill files under ~/.claude/skills/)
source ~/.claude/skills/iotops-verify-pr-slack/slack-templates.md 2>/dev/null || true
# If sourcing fails (non-bash context), define slack_post inline (copy from slack-templates.md).

FIXER_ISSUE_URL="${MULTICA_APP_URL:-http://localhost:3001}/issues/${FIXER_ISSUE_ID}"

SCOUT_MSG=$(printf ':bug: *New IoTOps error detected*\n*Service:* `%s`  |  *Cluster:* %s\n*Exception:* `%s` at `%s:%s`\n*Jira:* <%s|%s>  |  *Humio:* <%s|Open>  |  *Issue:* <%s|View in Multica>' \
  "$SERVICE" "$CLUSTER" "$EXC" "$FILE" "$LINE" "$JIRA_URL" "$JIRA_KEY" "$HUMIO_LINK" "$FIXER_ISSUE_URL")

SLACK_RESP=$(slack_post "$SCOUT_MSG")
SLACK_THREAD_TS=$(echo "$SLACK_RESP" | jq -r '.ts // empty')

if [ -n "$SLACK_THREAD_TS" ]; then
  multica issue comment add "$FIXER_ISSUE_ID" \
    --content "<!-- slack-thread-ts: ${SLACK_THREAD_TS} -->"
fi
```

The `slack_post` helper in `slack-templates.md` falls back to a plain webhook post
when the bot token is missing — the Fixer/Verifier will simply post new top-level
messages instead of thread replies in that case.

### 9. Close out

Use `multica issue comment add "$SCOUT_ISSUE_ID" --content-stdin` to post a summary comment on your own tick issue, then `multica issue status "$SCOUT_ISSUE_ID" done` (or `blocked` on total Humio failure). Both subcommands documented in the `multica-cli` skill's `issue-ops.md`.

Summary comment should include:
- Clusters scanned: 4
- Events found (total, per cluster)
- Distinct errors after fingerprint: N
- Fixer issues created: M (list their IDs, so the chain is browsable)
- Skipped (already open): K
- Skipped (code 404): L

## Do not

- Do **not** propose fixes or write patches — that's the Fixer's job. Your output is classification + context.
- Do **not** create a Fixer child if the JIRA search found an existing ticket **AND** there's an open Fixer/Verifier chain already in multica (dedup via step 6).
- Do **not** include more than 200 lines of ±10-min context per error — the Fixer can query back for more if needed.
