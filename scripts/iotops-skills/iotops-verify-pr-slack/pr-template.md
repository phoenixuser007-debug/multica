# Stash draft PR — request shape

## Endpoint

```
POST https://stash.arubanetworks.com/rest/api/1.0/projects/GVT/repos/<repo>/pull-requests
Authorization: Bearer $BITBUCKET_STASH_TOKEN
Content-Type: application/json
```

## Request body

```json
{
  "title": "<JIRA_KEY>: <one-line root cause>",
  "description": "<markdown body, see below>",
  "fromRef": {
    "id": "refs/heads/bugfix/<JIRA_KEY>-<slug>",
    "repository": { "slug": "<repo>", "project": { "key": "GVT" } }
  },
  "toRef": {
    "id": "refs/heads/devel",
    "repository": { "slug": "<repo>", "project": { "key": "GVT" } }
  },
  "state": "OPEN",
  "open": true,
  "closed": false,
  "draft": true
}
```

## Title rules

- Must start with the Jira key: `CNX-XXXXXX:` — Bitbucket Automation uses the prefix to link the build to the ticket.
- Keep the rest under 72 chars.
- Examples: `CNX-240415: fix npe in ktable join on state-processor`

## Description template (Markdown)

The PR description must be a **human-readable narrative**. Reviewers can already
see the diff in the Files-changed tab — do **not** repeat it in the description.

```markdown
## What was happening
{{2–4 sentences describing the production symptom in plain English: which
service, which exception, what user-visible or system-visible impact, and why
the existing code allowed it. No stack traces, no log dumps — explain the
behaviour, not the artifact.}}

## What this PR does
{{2–4 sentences describing the fix in plain English: what the new code does
differently, where the change lives (file + class + method, no diff), and why
this is the minimal correct change rather than a workaround.}}

## Why this prevents recurrence
{{2–4 sentences: what invariant the fix establishes, which conditions are now
handled (null/empty/transient), and the regression test that locks the
behaviour in. Mention the new JUnit test by class + method name so reviewers
know what to look for in the Files tab.}}

## Validation
- `compile.sh` → PASS (dev-env-dev-1)
- `unit-tests.sh` → PASS — new test `{{test_class}}.{{test_method}}` was red before fix, green after
- `lint-source-code.sh` → PASS (dev-env-dev-1)
- `lint-k8s-objects.sh` → PASS (dev-env-dev-1)

## Links
- JIRA: {{JIRA_URL}}
- Humio: {{humio_deep_link}}

---
*Draft PR raised by the IoTOps Error Monitor multi-agent chain. A human will review + merge.*
```

### Tone rules

- Write for a senior engineer who has not seen the Jira ticket. They should
  understand the bug, the fix, and the safeguard in under a minute.
- No bullet lists of "changes made" — the diff already lists those.
- No copy-pasted stack traces, log lines, or `diff` blocks.
- Past tense for "what was happening", present tense for "what this PR does",
  future tense for "why this prevents recurrence".
- Use backticks for class/method/file names; do not use code fences for prose.

## After creation

The response has `links.self[0].href` (browser URL) and `id` (integer ID used for activity polling).

## Adding reviewers (optional)

To add a default reviewer group:

```bash
curl -sf -X PUT \
  -H "Authorization: Bearer $BITBUCKET_STASH_TOKEN" \
  -H "Content-Type: application/json" \
  "https://stash.arubanetworks.com/rest/api/1.0/projects/GVT/repos/$REPO/pull-requests/$PR_ID/participants/$REVIEWER_USER" \
  -d '{"role":"REVIEWER"}'
```

Default reviewer mapping is workspace-specific; leave empty if unsure — the automation will assign defaults via the branch-permission rules.
