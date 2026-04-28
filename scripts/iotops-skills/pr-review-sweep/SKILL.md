---
name: pr-review-sweep
description: "Sweep all open PRs raised by the IoTOps Fixer or Bridge-5G Fixer (≥6h old) for unresolved human review comments and dispatch a child multica issue per PR to the Java Dev Agent to resolve them. Use when an autopilot 'Sweep fixer PR reviews' issue fires, or when a fixer/verifier hands off a '## PR Review Sweep' issue listing explicit PR URLs."
allowed-tools: ["Read", "Bash", "Grep"]
---

# PR Review Sweep

Shared by **both** the IoTOps and Bridge-5G fixer chains. After a Fixer/Verifier raises a PR, human reviewers need ~6 hours to leave feedback. This skill runs **after that window**, scoops up review comments per PR, and hands each PR's bundle to the **Java Dev Agent** for follow-through.

## When this skill fires

The current issue's title starts with `Sweep fixer PR reviews` (created by the multica autopilot — see **Wiring** below) **or** the description contains a `## PR Review Sweep` heading with an explicit list of PR URLs.

## Prerequisites

- `BITBUCKET_STASH_TOKEN` — Stash PAT (repo-read)
- `MULTICA_API_URL`, `MULTICA_TOKEN`, `MULTICA_WORKSPACE_ID`
- `STASH_URL` — Stash base URL (e.g. `https://stash.arubanetworks.com`)
- `VERIFIER_AGENTS` — comma-separated multica agent names whose `## Fix Ready` issues identify a chain-raised PR (default `IoTOps Verifier`; extend to `IoTOps Verifier,Bridge-5G Verifier` once that chain has its own verifier). Single env var, not one-per-chain.
- `pr-cli` and `multica-cli` skills attached
- The standard pre-cloned checkouts at `/root/dev-env/ws/<repo>` (same as both Fixers; ii-ae repos use `/root/central/<repo>` per the multica CVE conventions)

## Steps

### 1. List eligible PRs

If the description has an explicit `**PR**:` list, use that. Otherwise enumerate **multica issues** assigned to every agent in `$VERIFIER_AGENTS` (the chain's record of "a PR was raised"), parse each issue's `## Fix Ready` description for `**Service**`, `**Branch**`, `**Jira Key**`, then look up the matching open PR in Stash by branch. Apply the 6-hour cutoff on the Stash PR's `createdDate`.

This is tighter than filtering Stash by author: only PRs the chain itself raised get picked up, never a manual PR you opened by hand. Exact pipeline lives in [comment-fetch.md](comment-fetch.md).

Output: TSV at `/tmp/pr-eligible.tsv` with columns `PR_ID  REPO  PR_URL  BRANCH  JIRA_KEY`.

### 2. Fetch review comments per PR

For each row, switch into the repo's host checkout (`/root/dev-env/ws/$REPO` for gvt repos, `/root/central/$REPO` for `ii-ae-*` repos), check out `$BRANCH`, run `pr-cli comments --output json`, and filter to comments that have:
- a non-null `commentAnchor` (file/line review comments — not general PR comments)
- `resolved != true`

There is **no** author-based filter. The `commentAnchor != null` check already excludes the only kind of comment the fixer/verifier ever leaves (general status updates have no anchor), and an author filter would silently drop legitimate review comments in setups where the chain's Stash identity is also a real human reviewer.

Drop PRs that end up with zero such comments — leave them alone. See [comment-fetch.md](comment-fetch.md) for the jq filter and the orphaned-comment-after-force-push edge case.

### 3. Dedupe — skip PRs already handed off

```bash
multica issue search "Address review comments on $REPO PR #$PR_ID" --output json \
  | jq 'map(select(.status != "done" and .status != "cancelled")) | length'
```

Non-zero → a Java-Dev-Agent issue is already in flight for this PR. Skip it and add a one-line note on the sweep issue.

### 4. Render the Java-Dev-Agent payload

Build `/tmp/javadev-${PR_ID}.md` from the template in [javadev-payload.md](javadev-payload.md). It carries PR/branch/Jira metadata, a `## Review Comments` block (one entry per thread, `file:line — author` heading + body + reply chain), and a `## What you must do` block prescribing the address → `ctl` → push-same-branch → `pr-cli reply` workflow.

### 5. Dispatch the child issue (UUID PUT — not `--assignee` by name)

`--assignee "Java Dev Agent"` is **ambiguous** in this workspace — multica's name resolver also matches `copilot java dev agent`. Always create with no assignee, then PUT the resolved UUID:

```bash
NEW_ID=$(multica issue create \
  --title "$JIRA_KEY: Address review comments on $REPO PR #$PR_ID" \
  --description "$(cat /tmp/javadev-${PR_ID}.md)" \
  --parent "$THIS_ISSUE_ID" \
  --priority medium --status todo \
  --output json | jq -r '.id')

TOKEN=$(jq  -r '.token'        /root/.multica/config.json)
WSID=$(jq   -r '.workspace_id' /root/.multica/config.json)
SERVER=$(jq -r '.server_url'   /root/.multica/config.json)
AGENT_ID=$(curl -sf -H "Authorization: Bearer $TOKEN" \
  "$SERVER/api/agents?workspace_id=$WSID" \
  | jq -r '.[] | select(.name == "Java Dev Agent") | .id')

curl -sf -X PUT \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Workspace-ID: $WSID" \
  -H "Content-Type: application/json" \
  -d "{\"assignee_type\":\"agent\",\"assignee_id\":\"$AGENT_ID\"}" \
  "$SERVER/api/issues/$NEW_ID" >/dev/null
```

### 6. Close out

```bash
multica issue comment add "$THIS_ISSUE_ID" --content-stdin <<EOF
Sweep complete. Dispatched <N> Java-Dev-Agent issue(s):
- <PR_URL_1> → <child_issue_url_1>
- <PR_URL_2> → <child_issue_url_2>
EOF
multica issue status "$THIS_ISSUE_ID" done
```

## Wiring — the multica autopilot

This skill is intended to fire on a recurring 6-hour cadence and is attached to a single shared agent named **PR Review Sweep** that covers both fixer chains. Set it up **once** with `multica autopilot`, not the external `/schedule` skill:

```bash
AID=$(multica autopilot create \
  --title "PR Review Sweep" \
  --mode create_issue \
  --agent "PR Review Sweep" \
  --priority low \
  --issue-title-template "Sweep fixer PR reviews ({{date}})" \
  --description "For every multica issue assigned to an agent in \$VERIFIER_AGENTS (default IoTOps Verifier), look up the matching Stash PR; if it is open, ≥6h old, and has unresolved human review comments, dispatch a child issue to the Java Dev Agent. Skill: pr-review-sweep." \
  --output json | jq -r '.id')

multica autopilot trigger-add "$AID" --cron "0 */6 * * *" --timezone UTC
```

The skill's per-PR ≥ 6h cool-down is enforced internally (Step 1 filter), so a missed cron tick is harmless — the next tick still picks up PRs that crossed the threshold. One agent + one autopilot covers every chain whose verifier name is in `$VERIFIER_AGENTS`; if a Bridge-5G Verifier (or any new verifier) is added later, append its name to that env var — no skill changes needed.

## Do not

- Do **not** reply to PR comments yourself — only the Java Dev Agent talks back to reviewers.
- Do **not** auto-resolve threads — humans decide when their feedback is addressed.
- Do **not** filter comments by author. The `commentAnchor != null` check already excludes the only kind of comment the chain ever posts.
- Do **not** use `--assignee "Java Dev Agent"`. Resolve the UUID and PUT it, per Step 5.
- Do **not** dispatch a duplicate child issue when a non-terminal one already exists for the same PR (Step 3 dedup).
- Do **not** sweep manual / non-chain PRs. The Verifier-issue source in Step 1 already excludes them; do not re-introduce a Stash author filter.
