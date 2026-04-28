#!/usr/bin/env bash
# setup-iotops-error-agents.sh
#
# Provisions the 5 skills + 4 agents that make up the IoTOps error-remediation
# chain in multica, plus the shared PR Review Sweep agent that runs across
# both the IoTOps and Bridge-5G fixer chains. Runs in three phases:
#
#   1. For each skill under scripts/iotops-skills/, POST /api/skills with
#      content=SKILL.md and files=[everything else in the dir]. Also imports
#      scripts/iotops-skills/iotops-error-pipeline-debug (a wrapper around the
#      8 stage SKILL.md files from $IOTOPS_MONITOR_DIR/skills/). New: also
#      provisions scripts/iotops-skills/pr-review-sweep.
#
#   2. Create 4 public agents (Scout, Fixer, Verifier, PR Review Sweep) via
#      POST /api/agents and attach their skills via PUT /api/agents/{id}/skills.
#      The PR Review Sweep agent additionally requires the workspace's existing
#      pr-cli and multica-cli skills — looked up by name and attached.
#
#   3. (optional, --with-autopilot) Wire a recurring autopilot for the PR
#      Review Sweep agent at cron "0 */6 * * *" via the local `multica` CLI.
#      Skipped if the CLI is missing — the script still prints the commands
#      to run by hand.
#
# Usage:
#   ./scripts/setup-iotops-error-agents.sh \
#     --api-url      http://localhost:8090 \
#     --workspace    <workspace-id> \
#     --token        <bearer-token> \
#     --runtime-id   <runtime-uuid> \
#     --monitor-dir  /root/experiment/iotops-error-monitor \
#     --with-autopilot
#
# Options:
#   --dry-run         Print requests that would be sent without executing them.
#   --skip-existing   Reuse existing skills/agents whose names already match.
#   --with-autopilot  After Phase 2, create the PR Review Sweep autopilot via
#                     the local `multica` CLI (cron "0 */6 * * *" UTC).
#
# Env-var equivalents: MULTICA_API_URL, MULTICA_WORKSPACE_ID, MULTICA_TOKEN,
# MULTICA_RUNTIME_ID, IOTOPS_MONITOR_DIR.

set -euo pipefail

# ── Defaults ────────────────────────────────────────────────────────────────

API_URL="${MULTICA_API_URL:-http://localhost:8090}"
WORKSPACE_ID="${MULTICA_WORKSPACE_ID:-}"
TOKEN="${MULTICA_TOKEN:-}"
RUNTIME_ID="${MULTICA_RUNTIME_ID:-}"
MONITOR_DIR="${IOTOPS_MONITOR_DIR:-/root/experiment/iotops-error-monitor}"
DRY_RUN=false
SKIP_EXISTING=false
WITH_AUTOPILOT=false

# ── Argument parsing ─────────────────────────────────────────────────────────

while [[ $# -gt 0 ]]; do
  case "$1" in
    --api-url)        API_URL="$2";        shift 2 ;;
    --workspace)      WORKSPACE_ID="$2";   shift 2 ;;
    --token)          TOKEN="$2";          shift 2 ;;
    --runtime-id)     RUNTIME_ID="$2";     shift 2 ;;
    --monitor-dir)    MONITOR_DIR="$2";    shift 2 ;;
    --dry-run)        DRY_RUN=true;        shift   ;;
    --skip-existing)  SKIP_EXISTING=true;  shift   ;;
    --with-autopilot) WITH_AUTOPILOT=true; shift   ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$WORKSPACE_ID" || -z "$TOKEN" || -z "$RUNTIME_ID" ]]; then
  echo "Error: --workspace, --token, and --runtime-id are required." >&2
  echo "       (or set MULTICA_WORKSPACE_ID / MULTICA_TOKEN / MULTICA_RUNTIME_ID)" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_SRC_DIR="$SCRIPT_DIR/iotops-skills"

if [[ ! -d "$SKILL_SRC_DIR" ]]; then
  echo "Error: skill source dir not found: $SKILL_SRC_DIR" >&2
  exit 1
fi

AUTH_HEADERS=(
  -H "Authorization: Bearer $TOKEN"
  -H "X-Workspace-ID: $WORKSPACE_ID"
  -H "Content-Type: application/json"
)

# ── Helpers ──────────────────────────────────────────────────────────────────

api_call() {
  local method="$1"; shift
  local path="$1"; shift
  if $DRY_RUN; then
    echo "[dry-run] $method $API_URL$path" >&2
    [[ $# -gt 0 ]] && echo "  body: $*" >&2
    return 0
  fi
  curl -sf "${AUTH_HEADERS[@]}" -X "$method" "$API_URL$path" "$@"
}

# Emit a JSON body for POST /api/skills from a directory.
# Args: $1 = skill dir (contains SKILL.md + optional extra files).
#       $2 = (optional) extra files dir to merge (for iotops-error-pipeline-debug).
build_skill_body() {
  local dir="$1"
  local extra_dir="${2:-}"

  local skill_md="$dir/SKILL.md"
  if [[ ! -f "$skill_md" ]]; then
    echo "Missing SKILL.md in $dir" >&2
    return 1
  fi

  # Extract frontmatter name + description.
  local name description
  name=$(awk '/^---$/{f++;next} f==1 && /^name:/ {sub(/^name:[ \t]*/,""); print; exit}' "$skill_md")
  description=$(awk '/^---$/{f++;next} f==1 && /^description:/ {sub(/^description:[ \t]*"?/,""); sub(/"$/,""); print; exit}' "$skill_md")
  [[ -z "$name" ]] && { echo "Skill at $dir missing 'name' in frontmatter" >&2; return 1; }

  # Collect non-SKILL.md files under $dir as files[] entries.
  local -a file_args=()
  while IFS= read -r -d '' f; do
    local rel="${f#$dir/}"
    file_args+=(--arg "path_$(echo "$rel" | tr -c 'a-zA-Z0-9' '_')" "$rel")
    file_args+=(--arg "content_$(echo "$rel" | tr -c 'a-zA-Z0-9' '_')" "$(cat "$f")")
  done < <(find "$dir" -type f ! -name 'SKILL.md' -print0 2>/dev/null)

  # If extra_dir is set (iotops-error-pipeline-debug case), include every
  # SKILL.md under it with path = "<last-dir-component>/SKILL.md".
  if [[ -n "$extra_dir" && -d "$extra_dir" ]]; then
    while IFS= read -r -d '' f; do
      local rel
      rel="$(basename "$(dirname "$f")")/SKILL.md"
      file_args+=(--arg "path_$(echo "$rel" | tr -c 'a-zA-Z0-9' '_')" "$rel")
      file_args+=(--arg "content_$(echo "$rel" | tr -c 'a-zA-Z0-9' '_')" "$(cat "$f")")
    done < <(find "$extra_dir" -mindepth 2 -maxdepth 2 -name 'SKILL.md' -print0)
  fi

  # Build the JSON via a python helper to avoid shell-quoting headaches.
  python3 - "$name" "$description" "$skill_md" "$dir" "$extra_dir" <<'PYEOF'
import json, os, sys
name, desc, skill_md, dir_, extra_dir = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4], sys.argv[5]
with open(skill_md) as f:
    content = f.read()
files = []
for root, _, names in os.walk(dir_):
    for n in names:
        if n == "SKILL.md":
            continue
        abs_p = os.path.join(root, n)
        rel_p = os.path.relpath(abs_p, dir_)
        with open(abs_p) as fp:
            files.append({"path": rel_p, "content": fp.read()})
if extra_dir and os.path.isdir(extra_dir):
    for sub in sorted(os.listdir(extra_dir)):
        sub_path = os.path.join(extra_dir, sub, "SKILL.md")
        if os.path.isfile(sub_path):
            with open(sub_path) as fp:
                files.append({"path": f"{sub}/SKILL.md", "content": fp.read()})
print(json.dumps({
    "name": name,
    "description": desc,
    "content": content,
    "config": {},
    "files": files,
}))
PYEOF
}

# Look up an existing skill ID by name. Returns empty string if not found.
find_skill_by_name() {
  local name="$1"
  if $DRY_RUN; then echo ""; return; fi
  curl -sf "${AUTH_HEADERS[@]}" "$API_URL/api/skills" \
    | python3 -c "import sys,json; d=json.load(sys.stdin);
skills=d if isinstance(d,list) else d.get('skills',[]);
[print(s['id']) for s in skills if s['name']=='$name']" | head -1
}

create_skill() {
  local dir="$1"
  local extra_dir="${2:-}"
  local body
  body=$(build_skill_body "$dir" "$extra_dir")
  local name
  name=$(printf '%s' "$body" | python3 -c "import sys,json;print(json.load(sys.stdin)['name'])")

  if $SKIP_EXISTING; then
    local existing
    existing=$(find_skill_by_name "$name")
    if [[ -n "$existing" ]]; then
      echo "  skill '$name' exists ($existing) — skipping" >&2
      echo "$existing"
      return
    fi
  fi

  if $DRY_RUN; then
    echo "[dry-run] POST $API_URL/api/skills" >&2
    printf '%s' "$body" | python3 -m json.tool | head -20 >&2
    echo "dry-run-skill-id-$name"
    return
  fi

  local resp
  resp=$(printf '%s' "$body" | curl -sf "${AUTH_HEADERS[@]}" -X POST "$API_URL/api/skills" --data-binary @-)
  local id
  id=$(printf '%s' "$resp" | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")
  echo "  created skill '$name' → $id" >&2
  echo "$id"
}

find_agent_by_name() {
  local name="$1"
  if $DRY_RUN; then echo ""; return; fi
  curl -sf "${AUTH_HEADERS[@]}" "$API_URL/api/agents" \
    | python3 -c "import sys,json; d=json.load(sys.stdin);
agents=d if isinstance(d,list) else d.get('agents',[]);
[print(a['id']) for a in agents if a['name']=='$name']" | head -1
}

create_agent() {
  local name="$1"
  local description="$2"
  local instructions="$3"

  if $SKIP_EXISTING; then
    local existing
    existing=$(find_agent_by_name "$name")
    if [[ -n "$existing" ]]; then
      echo "  agent '$name' exists ($existing) — skipping" >&2
      echo "$existing"
      return
    fi
  fi

  local body
  body=$(python3 -c "
import json,sys
print(json.dumps({
  'name': sys.argv[1],
  'description': sys.argv[2],
  'instructions': sys.argv[3],
  'runtime_id': sys.argv[4],
  'visibility': 'workspace',
  'max_concurrent_tasks': 4,
}))" "$name" "$description" "$instructions" "$RUNTIME_ID")

  if $DRY_RUN; then
    echo "[dry-run] POST $API_URL/api/agents" >&2
    printf '%s' "$body" | python3 -m json.tool >&2
    echo "dry-run-agent-id-$name"
    return
  fi

  local resp
  resp=$(printf '%s' "$body" | curl -sf "${AUTH_HEADERS[@]}" -X POST "$API_URL/api/agents" --data-binary @-)
  local id
  id=$(printf '%s' "$resp" | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")
  echo "  created agent '$name' → $id" >&2
  echo "$id"
}

attach_skills() {
  local agent_id="$1"
  shift
  local -a skill_ids=("$@")
  local body
  body=$(python3 -c "
import json,sys
print(json.dumps({'skill_ids': sys.argv[1:]}))" "${skill_ids[@]}")

  if $DRY_RUN; then
    echo "[dry-run] PUT $API_URL/api/agents/$agent_id/skills  body=$body" >&2
    return
  fi

  printf '%s' "$body" | curl -sf "${AUTH_HEADERS[@]}" -X PUT \
    "$API_URL/api/agents/$agent_id/skills" --data-binary @- > /dev/null
  echo "  attached ${#skill_ids[@]} skill(s) to agent $agent_id"
}

# ── Phase 1: skills ──────────────────────────────────────────────────────────

echo "Phase 1/2 — skills"

SCOUT_SKILL_ID=$(create_skill "$SKILL_SRC_DIR/iotops-scout")
FIXER_SKILL_ID=$(create_skill "$SKILL_SRC_DIR/iotops-fix-apply")
VERIFIER_SKILL_ID=$(create_skill "$SKILL_SRC_DIR/iotops-verify-pr-slack")
SWEEP_SKILL_ID=$(create_skill "$SKILL_SRC_DIR/pr-review-sweep")

# The engineer-facing bundle picks up the 8 stage SKILL.md files from the
# legacy iotops-error-monitor tree.  If the tree is missing we skip it with a
# warning rather than failing the whole script.
DEBUG_WRAPPER_DIR="$SKILL_SRC_DIR/iotops-error-pipeline-debug"
DEBUG_SOURCE_DIR="$MONITOR_DIR/skills"
if [[ -d "$DEBUG_SOURCE_DIR" && -f "$DEBUG_WRAPPER_DIR/SKILL.md" ]]; then
  DEBUG_SKILL_ID=$(create_skill "$DEBUG_WRAPPER_DIR" "$DEBUG_SOURCE_DIR")
else
  if [[ ! -d "$DEBUG_SOURCE_DIR" ]]; then
    echo "  iotops-error-monitor debug source not found at $DEBUG_SOURCE_DIR — skipping debug skill"
  fi
  if [[ ! -f "$DEBUG_WRAPPER_DIR/SKILL.md" ]]; then
    echo "  debug-skill wrapper SKILL.md missing at $DEBUG_WRAPPER_DIR — skipping debug skill"
  fi
  DEBUG_SKILL_ID=""
fi

# pr-cli and multica-cli are workspace-wide skills provisioned outside this
# script (they're used by every IoTOps and CVE agent). The PR Review Sweep
# agent needs both attached. Look up their existing IDs by name; fail fast if
# they aren't in the workspace yet — we don't silently provision a half-wired
# Sweep agent.
if $DRY_RUN; then
  PR_CLI_SKILL_ID="dry-run-skill-id-pr-cli"
  MULTICA_CLI_SKILL_ID="dry-run-skill-id-multica-cli"
else
  PR_CLI_SKILL_ID=$(find_skill_by_name "pr-cli")
  MULTICA_CLI_SKILL_ID=$(find_skill_by_name "multica-cli")
  if [[ -z "$PR_CLI_SKILL_ID" || -z "$MULTICA_CLI_SKILL_ID" ]]; then
    echo "Error: PR Review Sweep needs the existing 'pr-cli' and 'multica-cli' skills to be present in the workspace." >&2
    [[ -z "$PR_CLI_SKILL_ID" ]]      && echo "       'pr-cli' skill not found." >&2
    [[ -z "$MULTICA_CLI_SKILL_ID" ]] && echo "       'multica-cli' skill not found." >&2
    echo "       Provision them first, then re-run this script." >&2
    exit 1
  fi
  echo "  resolved skill 'pr-cli'      → $PR_CLI_SKILL_ID"
  echo "  resolved skill 'multica-cli' → $MULTICA_CLI_SKILL_ID"
fi

# ── Phase 2: agents ──────────────────────────────────────────────────────────

echo ""
echo "Phase 2/2 — agents"

SCOUT_AGENT_ID=$(create_agent \
  "IoTOps Scout" \
  "Polls Humio for iotops-client-* runtime errors and hands each new one to the IoTOps Fixer as a child issue." \
  "You are the IoTOps Error Scout. Follow the iotops-scout skill exactly. Always dedup via multica issue search before creating a Fixer child issue.")

FIXER_AGENT_ID=$(create_agent \
  "IoTOps Fixer" \
  "Given a Scout-prepared error-context issue, performs RCA, generates a minimal fix, validates it in dev-env-dev-1, and hands off to the IoTOps Verifier." \
  "You are the IoTOps Fixer. Follow the iotops-fix-apply skill exactly. Never push a branch that failed dev-env compile or lint.")

VERIFIER_AGENT_ID=$(create_agent \
  "IoTOps Verifier" \
  "Opens the JIRA ticket, draft PR in Stash, polls Jenkins CI, and posts the final Slack notification for the IoTOps error chain." \
  "You are the IoTOps Verifier. Follow the iotops-verify-pr-slack skill exactly. Do not auto-merge PRs — a human reviews and merges.")

SWEEP_AGENT_ID=$(create_agent \
  "PR Review Sweep" \
  "Sweeps PRs raised by the IoTOps and Bridge-5G fixer chains every 6 hours; dispatches review-comment follow-up issues to the Java Dev Agent." \
  "You are the PR Review Sweep agent. Follow the pr-review-sweep skill exactly. Iterate \$FIXER_USERS to cover both the IoTOps and Bridge-5G fixer bots. Never reply to PR comments yourself — only the Java Dev Agent does.")

# Attach skills.
attach_skills "$SCOUT_AGENT_ID" "$SCOUT_SKILL_ID"
attach_skills "$FIXER_AGENT_ID" "$FIXER_SKILL_ID"
attach_skills "$VERIFIER_AGENT_ID" "$VERIFIER_SKILL_ID"
attach_skills "$SWEEP_AGENT_ID" "$SWEEP_SKILL_ID" "$PR_CLI_SKILL_ID" "$MULTICA_CLI_SKILL_ID"

# ── Phase 3: PR Review Sweep autopilot (optional) ───────────────────────────

SWEEP_AUTOPILOT_ID=""
SWEEP_AUTOPILOT_CRON='0 */6 * * *'
SWEEP_AUTOPILOT_TZ='UTC'

if $WITH_AUTOPILOT; then
  echo ""
  echo "Phase 3/3 — PR Review Sweep autopilot"

  if ! command -v multica >/dev/null 2>&1; then
    echo "  multica CLI not on PATH — skipping autopilot creation." >&2
    echo "  Run the equivalent commands by hand (printed in the summary below)." >&2
  elif $DRY_RUN; then
    echo "[dry-run] multica autopilot create --title 'PR Review Sweep' --agent 'PR Review Sweep' …"
    echo "[dry-run] multica autopilot trigger-add <id> --cron '$SWEEP_AUTOPILOT_CRON' --timezone $SWEEP_AUTOPILOT_TZ"
    SWEEP_AUTOPILOT_ID="dry-run-autopilot-id"
  else
    SWEEP_AUTOPILOT_ID=$(MULTICA_SERVER_URL="$API_URL" MULTICA_WORKSPACE_ID="$WORKSPACE_ID" \
      multica autopilot create \
        --title "PR Review Sweep" \
        --mode create_issue \
        --agent "PR Review Sweep" \
        --priority low \
        --issue-title-template "Sweep fixer PR reviews ({{date}})" \
        --description "Sweep all open PRs authored by the configured fixer bots (\$FIXER_USERS); for any PR ≥6h old with unresolved human review comments, dispatch a child issue to the Java Dev Agent. Skill: pr-review-sweep." \
        --output json \
      | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")

    MULTICA_SERVER_URL="$API_URL" MULTICA_WORKSPACE_ID="$WORKSPACE_ID" \
      multica autopilot trigger-add "$SWEEP_AUTOPILOT_ID" \
        --cron "$SWEEP_AUTOPILOT_CRON" \
        --timezone "$SWEEP_AUTOPILOT_TZ" \
        --label "every 6h" >/dev/null

    echo "  created autopilot 'PR Review Sweep' → $SWEEP_AUTOPILOT_ID (cron $SWEEP_AUTOPILOT_CRON $SWEEP_AUTOPILOT_TZ)"
  fi
fi

# ── Summary ──────────────────────────────────────────────────────────────────

echo ""
echo "Done."
echo ""
echo "  Scout agent:           $SCOUT_AGENT_ID"
echo "  Fixer agent:           $FIXER_AGENT_ID"
echo "  Verifier agent:        $VERIFIER_AGENT_ID"
echo "  PR Review Sweep agent: $SWEEP_AGENT_ID"
if [[ -n "$SWEEP_AUTOPILOT_ID" ]]; then
  echo "  Sweep autopilot:       $SWEEP_AUTOPILOT_ID"
fi
echo ""
echo "Export these before running setup-iotops-error-autopilot.sh:"
echo ""
echo "  export IOTOPS_SCOUT_AGENT_ID=$SCOUT_AGENT_ID"
echo "  export IOTOPS_FIXER_AGENT_ID=$FIXER_AGENT_ID"
echo "  export IOTOPS_VERIFIER_AGENT_ID=$VERIFIER_AGENT_ID"
echo "  export PR_REVIEW_SWEEP_AGENT_ID=$SWEEP_AGENT_ID"
echo ""
echo "PR Review Sweep agent runtime config (set via the agent's env in multica):"
echo "  FIXER_USERS         — comma-separated Stash logins of every fixer bot"
echo "                        (e.g. 'iotops-fixer-bot,bridge5g-fixer-bot')"
echo "  STASH_URL           — Stash base URL (e.g. https://stash.arubanetworks.com)"
echo "  BITBUCKET_STASH_TOKEN — repo-read PAT for the sweep account"
if ! $WITH_AUTOPILOT; then
  echo ""
  echo "To wire the recurring sweep (every 6h), re-run with --with-autopilot, or"
  echo "run by hand:"
  echo ""
  echo "  multica autopilot create --title 'PR Review Sweep' --mode create_issue \\"
  echo "    --agent 'PR Review Sweep' --priority low \\"
  echo "    --issue-title-template 'Sweep fixer PR reviews ({{date}})' \\"
  echo "    --description 'Sweep all open PRs authored by the configured fixer bots (\$FIXER_USERS); …'"
  echo "  multica autopilot trigger-add <id> --cron '$SWEEP_AUTOPILOT_CRON' --timezone $SWEEP_AUTOPILOT_TZ"
fi
