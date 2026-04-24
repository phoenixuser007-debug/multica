# `multica agent` and `multica skill` — agent-ready recipes

Agents rarely need to create or archive other agents — that's a human-admin operation. The useful subcommands are **lookups** (who is the Fixer agent? what skills are attached?) and **introspection** (what tasks am I currently running?).

## Resolve an agent UUID by name

```bash
# Handoff targets for the iotops chain
FIXER_ID=$(multica agent list --output json \
  | jq -r '.[] | select(.name == "IoTOps Fixer") | .id')
VERIFIER_ID=$(multica agent list --output json \
  | jq -r '.[] | select(.name == "IoTOps Verifier") | .id')
```

In most flows you don't need the UUID — `multica issue create --assignee "IoTOps Fixer"` resolves the name on the server. Use UUID lookup only when you need to filter tasks or skills by agent.

## See your own skills

```bash
# $AGENT_ID = your own ID; passed into task env as IOTOPS_<ROLE>_AGENT_ID, or
# look up by your known name.
multica agent get "$AGENT_ID" --output json | jq '.skills | map(.name)'
```

Sanity check before running a step that depends on an attached skill (e.g. the Verifier expects `pr-cli` + `jenkins-skill` + `jira-cli` to be present).

## List tasks (your own queue)

```bash
multica agent tasks "$AGENT_ID" --output json \
  | jq 'map(select(.status == "pending" or .status == "running"))'
```

## Skill introspection

```bash
# See all skills in the workspace (names + descriptions).
multica skill list

# Full content of one skill, including bundled files.
multica skill get "$SKILL_ID" --output json \
  | jq '{name, description, files: (.files|map(.path))}'
```

## Reading bundled files from your attached skill

Skills materialize to `.claude/skills/<skill-name>/` in your working directory at task-claim time. You don't need `multica skill get` to read them at runtime — the files are already on disk:

```bash
# Inside your agent task:
cat .claude/skills/iotops-scout/humio-lsql.md
cat .claude/skills/multica-cli/issue-ops.md
```

Use `multica skill get` only when you need to **verify** the server-side content matches what's on disk (rare — usually a debugging step).

## Things not to do

- Do **not** `multica agent create` / `agent archive` from inside an agent task. These mutate workspace membership and should remain a human-admin operation.
- Do **not** `multica skill update` / `skill files upsert` from inside an agent — let the owning human control skill content. (The exception is a dedicated "skill-author" agent running the `write-a-skill` skill on an issue explicitly asking for it.)
- Do **not** loop over `multica agent list` to poll for another agent's completion — use `multica issue runs <id>` on the child issue instead; it's scoped and cheaper.
