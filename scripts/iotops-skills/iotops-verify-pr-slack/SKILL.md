---
name: iotops-verify-pr-slack
description: "Given a Fixer-pushed validated bugfix branch (Jira key + branch already exist from Scout/Fixer), open the draft PR against devel via the pr-cli skill with the canonical '<JIRA_KEY>: <summary>' title, poll Jenkins CI via jenkins-skill, then POST the final outcome to /api/slack/iotops-error-done. Triggers on issues whose title starts with a JIRA key and description contains a '## Fix Ready' heading."
allowed-tools: ["Read", "Bash", "Grep"]
---

# IoTOps Verifier

You are the **final link**. Scout already opened the Jira ticket. Fixer wrote the failing test, fixed the code, ran `ctl`, and pushed the branch. Your job: draft PR → CI → Slack.

## When this skill fires

Issue title starts with a JIRA key (e.g. `CNX-240415: Raise PR for …`) and description contains a `## Fix Ready` heading produced by the Fixer.

## Prerequisites

- `BITBUCKET_STASH_TOKEN` — Stash PAT
- `JIRA_TOKEN`, `JIRA_USERNAME` — for the `In Review` transition + comment
- `SLACK_WEBHOOK_URL` — you post to this directly (see `slack-templates.md`). No multica server endpoint is involved.
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
- `**Fixer Issue**:` → `$FIXER_ISSUE_URL`
- The unified diff from the `## Patch` code block → `$PATCH`

### 2. Open the draft PR via `pr-cli`

Use the `pr-cli` skill (it knows the Stash API conventions for GVT repos). Title **must** start with the Jira key — Bitbucket Automation links the build to the ticket via the prefix.

```
Title: ${JIRA_KEY}: <root cause one-liner from the TDD Summary>
From:  refs/heads/${BRANCH}
To:    refs/heads/devel
Draft: true
```

PR description (markdown):

```markdown
## Summary
Auto-generated fix for **${JIRA_KEY}**.

## TDD Summary
<copy the bullet list from the Fixer's ## TDD Summary>

## Patch
```diff
<contents of $PATCH>
```

## Links
- Jira: ${JIRA_URL}
- Humio: ${HUMIO_LINK}
- Multica: ${FIXER_ISSUE_URL}

---
*Draft PR raised by the IoTOps Error Monitor agent chain. A human will review + merge.*
```

Save the returned PR URL → `$PR_URL` and PR id → `$PR_ID`.

### 3. Transition Jira to **In Review** + comment

Use the `jira-cli` skill:
- Add a comment: `PR: $PR_URL`
- Transition: `In Review` (discover the transition ID if not cached — CVE flow has already proven this pattern).

### 4. Poll Jenkins CI via `jenkins-skill`

Wait up to 15 minutes for Bitbucket-Automation-triggered Jenkins builds on the new PR to finish. Use the `jenkins-skill` (which already knows how to fetch build status from the PR activities endpoint). Outcomes:
- All `SUCCESSFUL` → `fix_success`
- Any `FAILED` → `ci_failed` (stop polling immediately)
- Still `INPROGRESS` after 15 min → `ci_timeout`

### 5. Post to Slack (direct webhook)

Use the `post_slack` shell helper from `slack-templates.md` — it wraps a
plain `curl` to `$SLACK_WEBHOOK_URL`, no multica-backend call. One of three
verifier outcomes:
- `fix_success` (CI all green) → `:bug: IoTOps auto-fix raised for <service>`
- `ci_failed` (any build failed) → `:warning: CI FAILED on auto-fix PR ...`
- `ci_timeout` (15 min, still running) → `:hourglass: CI still running ...`

Full invocation examples in `slack-templates.md`.

If `$SLACK_WEBHOOK_URL` is unset, log a warning and continue — do not fail
the Verifier issue over a missing webhook.

### 6. Close out

- Final comment on this Verifier issue: `$JIRA_URL`, `$PR_URL`, CI result, links to Fixer + Scout parents.
- Set this issue to `done`.
- If CI failed or timed out, also set it to `blocked` instead of `done` so it surfaces for human review — the retry scheduler will pick it up.

## Do not

- Do **not** open a **second** Jira ticket — Scout already did it. Your job is to transition, not create.
- Do **not** auto-merge the PR. A human reviews.
- Do **not** bypass `pr-cli` / `jenkins-skill` / `jira-cli` — they encode the team's conventions (title prefix, draft flag, default reviewers, CI build-key naming). Using raw curl skips those and produces inconsistent PRs.
