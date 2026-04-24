# Slack notifications — posted directly to the webhook

The Verifier (and the Fixer, for validation failures) post to Slack by
curling `$SLACK_WEBHOOK_URL` directly. No multica server endpoint is
involved — everything lives in this skill.

## Env var

`SLACK_WEBHOOK_URL` must be on the runtime host (same env var the CVE flow
uses). If it's missing, log a warning and skip the Slack step — do not fail
the Verifier issue over a missing webhook.

## Outcome → message template

| Outcome | Who sends | Header |
|---|---|---|
| `fix_success` | Verifier | `:bug: IoTOps auto-fix raised for {service}` |
| `ci_failed` | Verifier | `:warning: CI FAILED on auto-fix PR for {service} — human review needed` |
| `ci_timeout` | Verifier | `:hourglass: CI still running after 15 min on auto-fix PR for {service}` |
| `validation_failed` | Fixer | `:warning: Auto-fix FAILED compile/lint in dev-env for {service} — no PR raised` |
| `external` | Scout or Fixer | `:warning: External/upstream error for {service} — manual investigation` |

## Payload body

All outcomes use the same Slack block-kit shape. Fill only the fields that
apply (omit blank lines — the rendering drops empty `*...*` fields).

```bash
post_slack() {
  local header="$1"       # e.g. ":bug: IoTOps auto-fix raised for iotops-client-state-processor"
  local service="$2"
  local root_cause="$3"
  local jira_url="$4"     # may be empty
  local jira_key="$5"     # may be empty
  local pr_url="$6"       # may be empty
  local pr_title="$7"     # may be empty
  local humio_url="$8"    # may be empty

  if [ -z "$SLACK_WEBHOOK_URL" ]; then
    echo "warn: SLACK_WEBHOOK_URL not set — skipping Slack notify" >&2
    return 0
  fi

  local lines=""
  lines+="*Service:* \`${service}\`"
  [ -n "$root_cause" ] && lines+=$'\n'"*Root cause:* ${root_cause}"
  [ -n "$jira_key" ] && [ -n "$jira_url" ] && lines+=$'\n'"*Jira:* <${jira_url}|${jira_key}>"
  if [ -n "$pr_url" ]; then
    local pt="${pr_title:-View PR}"
    lines+=$'\n'"*PR:* <${pr_url}|${pt}>"
  fi
  [ -n "$humio_url" ] && lines+=$'\n'"*Humio:* <${humio_url}|Open in Humio>"

  local payload
  payload=$(jq -n \
    --arg txt "$header" \
    --arg body "$lines" \
    '{text:$txt, blocks:[{type:"section", text:{type:"mrkdwn", text:$body}}]}')

  curl -sf -X POST -H "Content-Type: application/json" \
    "$SLACK_WEBHOOK_URL" -d "$payload"
}
```

## Verifier usage (fix_success)

```bash
post_slack \
  ":bug: IoTOps auto-fix raised for \`$SERVICE\`" \
  "$SERVICE" "$ROOT_CAUSE_SHORT" \
  "$JIRA_URL" "$JIRA_KEY" \
  "$PR_URL" "${JIRA_KEY}: $ROOT_CAUSE_SHORT" \
  "$HUMIO_LINK"
```

## Verifier usage (ci_failed / ci_timeout)

```bash
post_slack \
  ":warning: CI FAILED on auto-fix PR for \`$SERVICE\` — human review needed" \
  "$SERVICE" "$ROOT_CAUSE_SHORT" \
  "$JIRA_URL" "$JIRA_KEY" \
  "$PR_URL" "$PR_TITLE" \
  "$HUMIO_LINK"
```

## Fixer usage (validation_failed)

```bash
post_slack \
  ":warning: Auto-fix FAILED compile/lint in dev-env for \`$SERVICE\` — no PR raised" \
  "$SERVICE" "$ROOT_CAUSE_SHORT" \
  "$JIRA_URL" "$JIRA_KEY" \
  "" "" \
  "$HUMIO_LINK"
```

## Why no server endpoint?

An earlier draft of this skill called a multica-backend endpoint
(`POST /api/slack/iotops-error-done`). That required Go code changes in the
server and added no value — agents already have `SLACK_WEBHOOK_URL` on the
runtime host, the message templates live here in the skill, and
`curl`-to-webhook is one function call. The server-side endpoint has been
removed. Post directly.
