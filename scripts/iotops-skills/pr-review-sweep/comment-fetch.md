# Fetching eligible PRs and review comments

Reference snippets for SKILL.md Steps 1 and 2.

## Step 1 — list eligible PRs

### Mode A: explicit PR list in the issue description

```bash
grep -E '^\*\*PR\*\*: ' /tmp/this-issue-description.md \
  | sed -E 's/^\*\*PR\*\*: //' \
  > /tmp/pr-urls.txt
```

For each URL, parse `(.+)/projects/([^/]+)/repos/([^/]+)/pull-requests/([0-9]+)$` to get `STASH_URL`, `PROJECT`, `REPO`, `PR_ID`. Then GET the PR detail to fill `BRANCH`, `JIRA_KEY` (from title prefix `CNX-...:`), and `createdDate`:

```bash
curl -sf -H "Authorization: Bearer $BITBUCKET_STASH_TOKEN" \
  "$STASH_URL/rest/api/1.0/projects/$PROJECT/repos/$REPO/pull-requests/$PR_ID" \
  | jq -r '[.id, .toRef.repository.slug, .links.self[0].href,
            .fromRef.displayId, (.title | capture("^(?<k>CNX-[0-9]+)").k // ""),
            .createdDate] | @tsv' \
  >> /tmp/pr-detail.tsv
```

Skip mode B and go to the cutoff filter below.

### Mode B: sweep — derive PRs from multica Verifier issues

The chain's source-of-truth for "a PR was raised" is the **Verifier issue** in multica: each one is created by the Fixer with a `## Fix Ready` description carrying `**Service**`, `**Branch**`, and `**Jira Key**`. Use those handoffs (not Stash author-filter) so a manual PR you opened by hand is never picked up.

`$VERIFIER_AGENTS` is the canonical comma-separated list of Verifier-like agent names (default `IoTOps Verifier`; extend to `IoTOps Verifier,Bridge-5G Verifier` if/when that chain adds its own verifier).

```bash
: > /tmp/verifier-handoffs.tsv
IFS=',' read -ra AGENTS <<<"${VERIFIER_AGENTS:-IoTOps Verifier}"

for AGENT in "${AGENTS[@]}"; do
  AGENT=$(echo "$AGENT" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
  [ -z "$AGENT" ] && continue

  # Skip cancelled — those handoffs were aborted. Keep done/in_review/blocked
  # because each represents a real PR that may still be open with comments.
  multica issue list --assignee "$AGENT" --limit 200 --output json \
    | jq -r '.issues[]
        | select(.status != "cancelled")
        | (.description // "")
        | capture("\\*\\*Service\\*\\*:\\s*`?(?<svc>[^`\n]+?)`?\\s*\\n[\\s\\S]*?\\*\\*Branch\\*\\*:\\s*`?(?<br>[^`\n]+?)`?\\s*\\n[\\s\\S]*?\\*\\*Jira Key\\*\\*:\\s*(?<jk>[A-Z]+-[0-9]+)") // empty
        | [.svc, .br, .jk] | @tsv' \
    >> /tmp/verifier-handoffs.tsv
done

# A Verifier issue can be re-run with a different Jira key against the same
# (service, branch) — keep the first row per (service, branch) so we don't
# query Stash twice for the same PR. The Jira-key column comes from whichever
# handoff appears first in the listing; downstream uses it only for the
# dispatched-issue title, so any matching key is fine.
awk -F'\t' '!seen[$1 FS $2]++' /tmp/verifier-handoffs.tsv > /tmp/verifier-handoffs.dedup
mv /tmp/verifier-handoffs.dedup /tmp/verifier-handoffs.tsv

# For each handoff, look up the matching open PR in Stash by branch.
: > /tmp/pr-detail.tsv
while IFS=$'\t' read -r SERVICE BRANCH JIRA_KEY; do
  case "$SERVICE" in ii-ae-*) PROJECT="II" ;; *) PROJECT="GVT" ;; esac

  curl -sf -H "Authorization: Bearer $BITBUCKET_STASH_TOKEN" \
    "$STASH_URL/rest/api/1.0/projects/$PROJECT/repos/$SERVICE/pull-requests?at=refs/heads/$BRANCH&state=OPEN&direction=OUTGOING&limit=1" \
    | jq -r --arg svc "$SERVICE" --arg jk "$JIRA_KEY" \
        '.values[]?
          | [.id, $svc, .links.self[0].href, .fromRef.displayId, $jk, .createdDate]
          | @tsv' \
    >> /tmp/pr-detail.tsv
done < /tmp/verifier-handoffs.tsv
```

`PROJECT` resolves to `II` for `ii-ae-*` repos and `GVT` for everything else, matching the multica CVE conventions. The Stash response carries `createdDate` (epoch ms) which feeds the 6-hour cutoff next.

### 6-hour cutoff

`createdDate` is **epoch milliseconds**. Filter to PRs older than 6h:

```bash
CUTOFF_MS=$(( $(date -u +%s) * 1000 - 21600000 ))
awk -v c="$CUTOFF_MS" -F'\t' '$6 != "" && $6+0 <= c' \
  /tmp/pr-detail.tsv \
  | cut -f1-5 > /tmp/pr-eligible.tsv
```

Resulting columns: `PR_ID  REPO  PR_URL  BRANCH  JIRA_KEY` — what Step 2 consumes.

## Step 2 — fetch + filter review comments per PR

For each row of `/tmp/pr-eligible.tsv`:

```bash
# Resolve the host checkout based on repo prefix (multica CVE conventions):
#   ii-ae-*  → /root/central/<repo>
#   default  → /root/dev-env/ws/<repo>
host_path() {
  case "$1" in
    ii-ae-*) echo "/root/central/$1" ;;
    *)       echo "/root/dev-env/ws/$1" ;;
  esac
}

while IFS=$'\t' read -r PR_ID REPO PR_URL BRANCH JIRA_KEY; do
  CHECKOUT=$(host_path "$REPO")
  cd "$CHECKOUT" || { echo "no checkout for $REPO at $CHECKOUT" >&2; continue; }
  git fetch -q origin "$BRANCH" 2>/dev/null || true
  git checkout -q "$BRANCH" 2>/dev/null || git checkout -q -B "$BRANCH" "origin/$BRANCH"

  pr-cli comments --output json > "/tmp/pr-${PR_ID}-comments.json"

  # Filter to unresolved file/line review comments. The `commentAnchor != null`
  # check already excludes the only kind of comment the fixer ever posts
  # (general PR status updates have no anchor), so we don't filter by author —
  # in some setups the fixer's Stash identity is the same human account that
  # also leaves real review feedback (e.g. naveen.u.holla on his own PRs), and
  # an author-based filter would silently drop those reviews.
  jq '
    [ .[]
      | select(.commentAnchor != null)
      | select(.resolved != true)
    ]
  ' "/tmp/pr-${PR_ID}-comments.json" > "/tmp/pr-${PR_ID}-review.json"

  COUNT=$(jq 'length' "/tmp/pr-${PR_ID}-review.json")
  if [ "$COUNT" -gt 0 ]; then
    printf '%s\t%s\t%s\t%s\t%s\t%s\n' \
      "$PR_ID" "$REPO" "$PR_URL" "$BRANCH" "$JIRA_KEY" "$COUNT" \
      >> /tmp/pr-needs-followup.tsv
  fi
done < /tmp/pr-eligible.tsv
```

### Edge cases

- **`pr-cli` returns `no open PR`** — the PR was merged or closed between Step 1 and Step 2. Skip silently; don't dispatch.
- **Orphaned comment after force-push** — Bitbucket retains the comment but `.commentAnchor.line` may point to a line that no longer exists. Treat orphaned comments as still-actionable (the reviewer's intent persists); the Java Dev Agent's `pr-cli reply` will close the loop.
- **Reply-only threads** — a comment with `parentId != null` is a reply, not a top-level review thread. The jq filter above includes them; render them under their parent in the payload (see `javadev-payload.md`), not as separate threads.
- **Author equals Stash bot but different display name** — Stash sometimes returns `author.name` (login) vs `author.displayName`. Filter on `author.name` only; the Fixer/Verifier configure the bot's login name, not its display name.

## Comment shape (input to the payload renderer)

Each item in `/tmp/pr-${PR_ID}-review.json` looks roughly like:

```json
{
  "id": 1234,
  "author": {"name": "alice.smith", "displayName": "Alice Smith"},
  "text": "This null guard is fine but consider using Optional.ofNullable here.",
  "commentAnchor": {
    "path": "src/main/java/com/aruba/iotops/StateProcessor.java",
    "line": 187,
    "lineType": "ADDED",
    "fileType": "TO"
  },
  "comments": [
    {"id": 1235, "author": {"name": "bob.jones"}, "text": "Agreed — Optional reads cleaner."}
  ],
  "resolved": false,
  "parentId": null
}
```

The renderer in Step 4 groups by `(commentAnchor.path, commentAnchor.line)` and inlines `comments[]` as the reply chain.
