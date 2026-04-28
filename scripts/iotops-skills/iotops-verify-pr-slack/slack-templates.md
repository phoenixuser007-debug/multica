# Slack notifications — posted directly to the webhook (or Bot API for threading)

The pipeline uses a **single Slack thread per error**, started by the Scout
and replied to by the Fixer (RCA) and Verifier (PR URL). Threading requires
the Slack Bot API rather than the incoming webhook, because webhooks do not
return a `ts` value needed to reply into a thread.

## Env vars

| Variable | Purpose |
|---|---|
| `SLACK_BOT_TOKEN` | Bot User OAuth Token (`xoxb-…`). Needs `chat:write` scope and must be added to the channel. Enables threading. |
| `SLACK_CHANNEL_ID` | Channel ID (e.g. `C0XXXXXXX`) where the error thread lives. |
| `SLACK_WEBHOOK_URL` | Fallback for one-shot posts when `SLACK_BOT_TOKEN` is not set. |

If `SLACK_BOT_TOKEN` + `SLACK_CHANNEL_ID` are set they take precedence; the
pipeline posts via `chat.postMessage` and can thread. Without them it falls
back to `SLACK_WEBHOOK_URL` for a flat (non-threaded) post. Do not fail the
agent run over a missing webhook.

## Core helper — `slack_post`

Paste this function at the top of any shell script that needs to post.
It returns the raw JSON response so callers can extract `.ts`.

```bash
slack_post() {
  local text="$1"
  local thread_ts="${2:-}"   # optional; set to reply inside an existing thread

  if [ -n "$SLACK_BOT_TOKEN" ] && [ -n "$SLACK_CHANNEL_ID" ]; then
    # Bot API — supports threading and returns ts
    local payload
    payload=$(jq -n \
      --arg ch  "$SLACK_CHANNEL_ID" \
      --arg txt "$text" \
      --arg ts  "$thread_ts" \
      '{ channel:$ch, text:$txt } + (if $ts != "" then {thread_ts:$ts} else {} end)')
    curl -sf -X POST \
      -H "Authorization: Bearer $SLACK_BOT_TOKEN" \
      -H "Content-Type: application/json" \
      "https://slack.com/api/chat.postMessage" \
      -d "$payload"
  elif [ -n "$SLACK_WEBHOOK_URL" ]; then
    # Fallback: webhook — no ts returned, no threading
    local payload
    payload=$(jq -n --arg t "$text" '{text:$t}')
    curl -sf -X POST -H "Content-Type: application/json" \
      "$SLACK_WEBHOOK_URL" -d "$payload" || true
    echo '{}'
  else
    echo "warn: no SLACK_BOT_TOKEN/SLACK_CHANNEL_ID or SLACK_WEBHOOK_URL set — skipping" >&2
    echo '{}'
  fi
}
```

## Helper — `extract_thread_ts`

Agents retrieve the thread timestamp from a multica issue's comments.
The Scout stores it using the sentinel `<!-- slack-thread-ts: <ts> -->`.

```bash
extract_thread_ts() {
  local issue_id="$1"
  multica issue comments list "$issue_id" --output json \
    | jq -r '.[].content' \
    | grep -oP '(?<=slack-thread-ts: )[\d.]+' \
    | head -1
}
```

If no sentinel is found (bot token was absent when Scout ran), `extract_thread_ts`
returns an empty string — pass it to `slack_post` and it will send a new top-level
message instead of a thread reply.

## Outcome → message template

| Outcome | Who sends | When |
|---|---|---|
| `error_found` | Scout | Error fingerprinted + Jira created + Fixer issue spawned |
| `rca_found` | Fixer | Proposed-fix comment posted (before TDD begins) |
| `fix_success` | Verifier | CI all green on draft PR |
| `ci_failed` | Verifier | Any CI build failed |
| `validation_failed` | Fixer | `ctl` compile/lint failed in dev-env |

The Verifier polls Jenkins every 10 minutes until CI reaches a terminal state — there is no `ci_timeout` outcome.

## Scout — initial thread message (`error_found`)

Post this immediately after creating the Fixer child issue. Capture the returned
`ts` and store it so the Fixer and Verifier can thread their replies.

```bash
# No multica link in Scout messages — Slack is for the team, multica audit
# trail stays inside multica. Jira + Humio are the only outbound links.
SCOUT_MSG=$(printf ':bug: *New IoTOps error detected*\n*Service:* `%s`  |  *Cluster:* %s\n*Exception:* `%s` at `%s:%s`\n*Jira:* <%s|%s>  |  *Humio:* <%s|Open>' \
  "$SERVICE" "$CLUSTER" "$EXC" "$FILE" "$LINE" "$JIRA_URL" "$JIRA_KEY" "$HUMIO_LINK")

SLACK_RESP=$(slack_post "$SCOUT_MSG")
SLACK_THREAD_TS=$(echo "$SLACK_RESP" | jq -r '.ts // empty')

# Store sentinel so Fixer / Verifier can thread their replies
if [ -n "$SLACK_THREAD_TS" ]; then
  multica issue comment add "$FIXER_ISSUE_ID" \
    --content "<!-- slack-thread-ts: ${SLACK_THREAD_TS} -->"
fi
```

## Fixer — RCA thread reply (`rca_found`)

Post this right after the proposed-fix comment is added to the multica issue.

```bash
SLACK_THREAD_TS=$(extract_thread_ts "$FIXER_ISSUE_ID")

RCA_MSG=$(printf ':mag: *RCA identified for `%s`*\n*Root cause:* %s\n*Approach:* %s\nImplementing fix (TDD red→green)…' \
  "$EXC" "$ROOT_CAUSE" "$APPROACH")

slack_post "$RCA_MSG" "$SLACK_THREAD_TS"
```

## Fixer — validation failure thread reply (`validation_failed`)

Post this if `ctl` fails before a Verifier is spawned.

```bash
SLACK_THREAD_TS=$(extract_thread_ts "$FIXER_ISSUE_ID")

VAL_MSG=$(printf ':warning: *Auto-fix FAILED compile/lint in dev-env*\n*Service:* `%s`  |  *Jira:* <%s|%s>\nNo PR raised — issue is blocked and needs human review.' \
  "$SERVICE" "$JIRA_URL" "$JIRA_KEY")

slack_post "$VAL_MSG" "$SLACK_THREAD_TS"
```

## Verifier — final thread reply

Post this after CI result is known (step 4 in the Verifier). The `FIXER_ISSUE_ID`
is extracted from the `**Fixer Issue**:` line in the Verifier issue description.

```bash
SLACK_THREAD_TS=$(extract_thread_ts "$FIXER_ISSUE_ID")

# fix_success
PR_MSG=$(printf ':white_check_mark: *Fix ready for review*\n*Service:* `%s`  |  *Jira:* <%s|%s>\n*Branch:* `%s`\n*PR:* <%s|View PR>  |  *CI:* :large_green_circle: SUCCESSFUL\n_A human must review and merge._' \
  "$SERVICE" "$JIRA_URL" "$JIRA_KEY" "$BRANCH" "$PR_URL")

# ci_failed
PR_MSG=$(printf ':warning: *CI FAILED on auto-fix PR*\n*Service:* `%s`  |  *Jira:* <%s|%s>\n*PR:* <%s|View PR>  |  *CI:* :red_circle: FAILED\nHuman review needed.' \
  "$SERVICE" "$JIRA_URL" "$JIRA_KEY" "$PR_URL")

slack_post "$PR_MSG" "$SLACK_THREAD_TS"
```

## Why no server endpoint?

An earlier draft of this skill called a multica-backend endpoint
(`POST /api/slack/iotops-error-done`). That required Go code changes in the
server and added no value — agents already have `SLACK_WEBHOOK_URL` on the
runtime host, the message templates live here in the skill, and
`curl`-to-webhook is one function call. The server-side endpoint has been
removed. Post directly.
