#!/usr/bin/env bash
# setup-iotops-error-agents.sh
#
# Provisions the 4 skills + 3 agents that make up the IoTOps error-remediation
# chain in multica. Runs in two phases:
#
#   1. For each skill under scripts/iotops-skills/, POST /api/skills with
#      content=SKILL.md and files=[everything else in the dir]. Also imports
#      scripts/iotops-skills/iotops-error-pipeline-debug (a wrapper around the
#      8 stage SKILL.md files from $IOTOPS_MONITOR_DIR/skills/).
#
#   2. Create 3 public agents (Scout, Fixer, Verifier) via POST /api/agents
#      and attach their skills via PUT /api/agents/{id}/skills.
#
# Usage:
#   ./scripts/setup-iotops-error-agents.sh \
#     --api-url      http://localhost:8090 \
#     --workspace    <workspace-id> \
#     --token        <bearer-token> \
#     --runtime-id   <runtime-uuid> \
#     --monitor-dir  /root/experiment/iotops-error-monitor
#
# Options:
#   --dry-run         Print requests that would be sent without executing them.
#   --skip-existing   Reuse existing skills/agents whose names already match.
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

# ── Argument parsing ─────────────────────────────────────────────────────────

while [[ $# -gt 0 ]]; do
  case "$1" in
    --api-url)      API_URL="$2";      shift 2 ;;
    --workspace)    WORKSPACE_ID="$2"; shift 2 ;;
    --token)        TOKEN="$2";        shift 2 ;;
    --runtime-id)   RUNTIME_ID="$2";   shift 2 ;;
    --monitor-dir)  MONITOR_DIR="$2";  shift 2 ;;
    --dry-run)      DRY_RUN=true;      shift   ;;
    --skip-existing) SKIP_EXISTING=true; shift ;;
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
      echo "  skill '$name' exists ($existing) — skipping"
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
  echo "  created skill '$name' → $id"
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
      echo "  agent '$name' exists ($existing) — skipping"
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
  echo "  created agent '$name' → $id"
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

# Attach skills.
attach_skills "$SCOUT_AGENT_ID" "$SCOUT_SKILL_ID"
attach_skills "$FIXER_AGENT_ID" "$FIXER_SKILL_ID"
attach_skills "$VERIFIER_AGENT_ID" "$VERIFIER_SKILL_ID"

# ── Summary ──────────────────────────────────────────────────────────────────

echo ""
echo "Done."
echo ""
echo "  Scout agent:     $SCOUT_AGENT_ID"
echo "  Fixer agent:     $FIXER_AGENT_ID"
echo "  Verifier agent:  $VERIFIER_AGENT_ID"
echo ""
echo "Export these before running setup-iotops-error-autopilot.sh:"
echo ""
echo "  export IOTOPS_SCOUT_AGENT_ID=$SCOUT_AGENT_ID"
echo "  export IOTOPS_FIXER_AGENT_ID=$FIXER_AGENT_ID"
echo "  export IOTOPS_VERIFIER_AGENT_ID=$VERIFIER_AGENT_ID"
