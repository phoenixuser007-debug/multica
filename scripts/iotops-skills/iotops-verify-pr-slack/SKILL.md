---
name: iotops-verify-pr-slack
description: "Given a Fixer-pushed validated bugfix branch (Jira key + branch already exist from Scout/Fixer), open the draft PR against devel via the pr-cli skill with the canonical '<JIRA_KEY>: <summary>' title, then poll Jenkins CI via jenkins-skill every 10 minutes until it finishes (no timeout — keep waiting), and finally post the threaded Slack reply with the result. Triggers on issues whose title starts with a JIRA key and description contains a '## Fix Ready' heading."
allowed-tools: ["Read", "Bash", "Grep"]
---

# IoTOps Verifier

You are the **final link**. Scout already opened the Jira ticket. Fixer wrote the failing test, fixed the code, ran `ctl`, and pushed the branch. Your job: draft PR → CI → Slack.

## When this skill fires

Issue title starts with a JIRA key (e.g. `CNX-240415: Raise PR for …`) and description contains a `## Fix Ready` heading produced by the Fixer.

## Prerequisites

- `BITBUCKET_STASH_TOKEN` — Stash PAT
- `JIRA_TOKEN`, `JIRA_USERNAME` — for the `In Review` transition + comment
- `SLACK_BOT_TOKEN` + `SLACK_CHANNEL_ID` — Bot API token and channel for threaded replies (optional; falls back to `SLACK_WEBHOOK_URL`)
- `SLACK_WEBHOOK_URL` — fallback for one-shot posts when bot token is not set
- `MULTICA_API_URL`, `MULTICA_TOKEN`, `MULTICA_WORKSPACE_ID`
- Skills attached: **`pr-cli`**, **`jenkins-skill`**, **`jira-cli`** — use those for their respective services, not raw curl.

## Steps

### 1. Parse the handoff

Extract from the `## Fix Ready` section:
- `**Service**:` → `$SERVICE`
- `**Branch**:` → `$BRANCH` (e.g. `bugfix/CNX-240415-null-pointer-exception`)
- `**Jira**:` → `$JIRA_URL`
- `**Jira Key**:` → `$JIRA_KEY`
- `**Humio**:` → `$HUMIO_LINK`
- `**Fixer Issue**:` → `$FIXER_ISSUE_URL` — also derive `$FIXER_ISSUE_ID` (the UUID at the end of the URL, or the `MCT-xxx` short ID, whichever form `multica issue comments list` accepts)
- The unified diff from the `## Patch` code block → `$PATCH`

### 2. Open the draft PR via `pr-cli`

Use the `pr-cli` skill (it knows the Stash API conventions for GVT repos). Title **must** start with the Jira key — Bitbucket Automation links the build to the ticket via the prefix.

```
Title: ${JIRA_KEY}: <root cause one-liner from the TDD Summary>
From:  refs/heads/${BRANCH}
To:    refs/heads/devel
Draft: true
```

PR description must be a **human-readable narrative**, not the diff. Use the
template in `pr-template.md` and fill the three narrative sections yourself —
do **not** paste `$PATCH` into the description, reviewers see the diff in the
Files-changed tab.

```markdown
## What was happening
<2–4 sentences. Plain English describing the production symptom: which service,
what exception, the user/system impact, and why the existing code allowed it.
No stack traces, no log lines.>

## What this PR does
<2–4 sentences. The fix in plain English: what the new code does differently,
the file/class/method it lives in (named, not pasted), and why this is the
minimal correct change.>

## Why this prevents recurrence
<2–4 sentences. The invariant the fix establishes, which conditions are now
handled (null/empty/transient), and the new JUnit test (by class + method name)
that locks the behaviour in.>

## Validation
- `compile.sh` → PASS (dev-env-dev-1)
- `unit-tests.sh` → PASS — new test `<TestClass>.<testMethod>` was red before fix, green after
- `lint-source-code.sh` → PASS (dev-env-dev-1)
- `lint-k8s-objects.sh` → PASS (dev-env-dev-1)

## Links
- Jira: ${JIRA_URL}
- Humio: ${HUMIO_LINK}

---
*Draft PR raised by the IoTOps Error Monitor agent chain. A human will review + merge.*
```

The three narrative sections are required — derive them from the Fixer's
`## Proposed Fix` comment (root cause + approach), the Fixer's `## TDD Summary`
(test class/method), and the Scout's `## Error Context` (production symptom).
Keep the prose tight; no bullet lists of "changes made", no code fences around
the narrative paragraphs. See `pr-template.md` for tone rules and a worked
example.

Save the returned PR URL → `$PR_URL` and PR id → `$PR_ID`.

#### 2a. If a PR already exists on the branch, rewrite its description

If `pr-cli` reports the PR already exists (Bitbucket returns 409 conflict or
`pr-cli` returns the existing PR id), **do not** open a duplicate. Instead:

1. Fetch the existing PR's current description.
2. Build the narrative description using the same template as above.
3. `PUT` the PR with the new description so it matches the template — even if
   the title is unchanged.

```bash
# After identifying $PR_ID for the existing PR
NEW_DESC=$(<build the narrative description as above>)
curl -sf -X PUT \
  -H "Authorization: Bearer $BITBUCKET_STASH_TOKEN" \
  -H "Content-Type: application/json" \
  "https://stash.arubanetworks.com/rest/api/1.0/projects/GVT/repos/${SERVICE}/pull-requests/${PR_ID}" \
  -d "$(jq -n --arg t "$PR_TITLE" --arg d "$NEW_DESC" --argjson v "$PR_VERSION" \
         '{title:$t, description:$d, version:$v}')"
```

(The `version` field is required by Stash's optimistic locking — fetch it from
the GET response first.) This step is idempotent: if the description already
matches the narrative template, the PUT is a no-op visually; if it still
contains the legacy `## Patch` diff block, the PUT replaces it. Always perform
this sync — older PRs raised by previous skill versions still carry the diff.

### 3. Transition Jira to **In Review** + comment

Use the `jira-cli` skill:
- Add a comment: `PR: $PR_URL`
- Transition: `In Review` (discover the transition ID if not cached — CVE flow has already proven this pattern).

### 4. Poll Jenkins CI via `jenkins-skill` — wait until CI finishes

Use the `jenkins-skill` (which already knows how to fetch build status from the PR activities endpoint). Poll **every 10 minutes** and **keep polling until every build reaches a terminal state** (`SUCCESSFUL` or `FAILED`).

Do **not** give up at 15 minutes. The Verifier issue must reflect the real CI outcome, so this step blocks until Jenkins is done.

```bash
while true; do
  STATUS=$(<query PR activities via jenkins-skill / pr-cli, return one of:
            SUCCESSFUL_ALL | FAILED_ANY | INPROGRESS>)
  case "$STATUS" in
    SUCCESSFUL_ALL) CI_OUTCOME=fix_success; break ;;
    FAILED_ANY)     CI_OUTCOME=ci_failed;   break ;;
    INPROGRESS)     sleep 600 ;;   # 10 minutes, then re-poll
  esac
done
```

Outcomes:
- All `SUCCESSFUL` → `CI_OUTCOME=fix_success`
- Any `FAILED` → `CI_OUTCOME=ci_failed` (stop polling immediately)

**Do not produce a `ci_timeout` outcome.** If you've been polling for hours, that is the correct behaviour — Jenkins is slow, the agent waits. Only stop on a terminal Jenkins state.

If you genuinely cannot reach Jenkins (e.g. token expired, network gone), post a comment on the Verifier issue explaining the auth/network problem, leave the issue in `in_progress` (not `blocked`), and exit. The retry scheduler will pick it up on the next cycle once the auth/network is fixed.

### 5. Post final Slack thread reply

Use the `slack_post` + `extract_thread_ts` helpers from `slack-templates.md`.
The `thread_ts` lives in the Fixer issue's comments (Scout stored it there).
If threading is unavailable, the helpers fall back to a flat webhook post.

```bash
source ~/.claude/skills/iotops-verify-pr-slack/slack-templates.md 2>/dev/null || true

# Read thread_ts from the Fixer issue
SLACK_THREAD_TS=$(extract_thread_ts "$FIXER_ISSUE_ID")

case "$CI_OUTCOME" in
  fix_success)
    PR_MSG=$(printf ':white_check_mark: *Fix ready for review*\n*Service:* `%s`  |  *Jira:* <%s|%s>\n*Branch:* `%s`\n*PR:* <%s|View PR>  |  *CI:* :large_green_circle: SUCCESSFUL\n_A human must review and merge._' \
      "$SERVICE" "$JIRA_URL" "$JIRA_KEY" "$BRANCH" "$PR_URL")
    ;;
  ci_failed)
    PR_MSG=$(printf ':warning: *CI FAILED on auto-fix PR*\n*Service:* `%s`  |  *Jira:* <%s|%s>\n*PR:* <%s|View PR>  |  *CI:* :red_circle: FAILED\nHuman review needed.' \
      "$SERVICE" "$JIRA_URL" "$JIRA_KEY" "$PR_URL")
    ;;
esac

slack_post "$PR_MSG" "$SLACK_THREAD_TS"
```

Full message templates also available in `slack-templates.md`. Do not fail
the Verifier issue over a missing webhook — log a warning and continue.

### 6. Close out

- Final comment on this Verifier issue: `$JIRA_URL`, `$PR_URL`, CI result, links to Fixer + Scout parents.
- Set this issue to `done` if `CI_OUTCOME=fix_success`.
- Set this issue to `blocked` only if `CI_OUTCOME=ci_failed` (so it surfaces for human review). **Never** set `blocked` because CI was slow — step 4 already waits as long as it takes.

## Do not

- Do **not** open a **second** Jira ticket — Scout already did it. Your job is to transition, not create.
- Do **not** auto-merge the PR. A human reviews.
- Do **not** bypass `pr-cli` / `jenkins-skill` / `jira-cli` — they encode the team's conventions (title prefix, draft flag, default reviewers, CI build-key naming). Using raw curl skips those and produces inconsistent PRs.
