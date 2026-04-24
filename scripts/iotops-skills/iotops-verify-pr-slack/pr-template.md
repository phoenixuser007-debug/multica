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

```markdown
## Summary
Auto-generated fix for **{{JIRA_KEY}}**.

## Root Cause
{{root_cause}}

## Affected File
```
{{affected_file_path}}
```

## Patch
```diff
{{git_diff_first_80_lines}}
```

## Validation
- `compile.sh` → PASS (dev-env-dev-1)
- `lint-source-code.sh` → PASS (dev-env-dev-1)

## Links
- JIRA: {{JIRA_URL}}
- Humio: {{humio_deep_link}}
- Multica (Fixer): {{fixer_issue_url}}
- Multica (Verifier): {{verifier_issue_url}}

---
*Draft PR raised by the IoTOps Error Monitor multi-agent chain. A human will review + merge.*
```

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
