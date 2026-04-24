---
name: iotops-fix-apply
description: "Given a Scout-prepared IoTOps error-context issue, propose a fix, then implement it using strict TDD: write a failing JUnit test that reproduces the exception first, run mvn test and confirm it FAILS (red), then implement the minimal fix, re-run and confirm GREEN, then run the ctl quality gate. On success, spawn a Verifier child issue. Triggers on issues whose description contains an '## Error Context' heading created by the IoTOps Scout."
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "Glob"]
---

# IoTOps Fixer (TDD mode)

You are the **second link** in the chain. Scout handed you a fully-contextualized issue with a Jira key already assigned. You **write the failing test first, then the code** — red → green → refactor → ctl. Only then do you hand off to the Verifier.

## When this skill fires

The current issue's description contains:
- `## Error Context` heading
- A `**Jira**:` line with the ticket URL (Scout created/reused it)
- A `**Source**:` line with `<service_repo>/<relative_path>:<line>`

## Prerequisites

- `BITBUCKET_STASH_TOKEN` — Stash PAT (repo-read + repo-write, SSH preferred for push)
- `MULTICA_API_URL`, `MULTICA_TOKEN`, `MULTICA_WORKSPACE_ID`, `IOTOPS_VERIFIER_AGENT_ID`
- `dev-env-dev-1` Docker container running on the host (same container CVE flow uses).
- The `ctl` skill is attached to this agent — use it via its own documented interface.

## Steps

### 1. Parse the handoff

Extract from the description:
- `**Cluster**:` → `$CLUSTER` (informational — for the Jira/PR comment)
- `**Service**:` → `$SERVICE` (also the affected repo name)
- `**Exception**:` → `$EXC`
- `**Source**:` → `$REL_PATH` and `$LINE_NO`
- `**Humio**:` → `$HUMIO_LINK`
- `**Jira**:` → `$JIRA_URL`, derive `$JIRA_KEY` (e.g. `CNX-240415`)
- Raw log from `## Error Log` code block
- Code snippet from `## Affected Source` code block

### 2. Propose the solution (post as a comment on this issue)

Before touching any code, reason through and post a comment:

```markdown
## Proposed Fix

**Root cause**: <one sentence — why `$EXC` happens at `$REL_PATH:$LINE_NO`>
**Approach**: <null guard / defensive retry / refactor / schema-version pin / ...>
**Minimal change**: <what lines are expected to change, and why>
**Test strategy**: <what the failing JUnit test will assert>
```

This forces an explicit plan before the TDD cycle.

### 3. Clone and branch

```bash
WORKDIR=$(mktemp -d -t iotops-fix-$SERVICE-XXXX)
cd "$WORKDIR"

# The repo lives at /root/dev-env/ws/<service> on the host (mirrors CVE path).
# Work inside a fresh shallow clone so the Fixer run is idempotent.
git clone --depth 1 --single-branch -b devel \
  "ssh://git@stash.arubanetworks.com/gvt/$SERVICE.git" repo
cd repo

SLUG=$(echo "$EXC" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | cut -c1-40)
BRANCH="bugfix/${JIRA_KEY}-${SLUG}"
git checkout -b "$BRANCH"
```

### 4. RED — write the failing test FIRST

This is the non-negotiable TDD step. Before editing the production code:

1. Locate the test source root: `src/test/java/...` mirroring the broken class's package.
2. Create a test file (e.g. `StateProcessorNullSafetyTest.java`) that reproduces the exception:
   - Use `@Test` with JUnit 5 (matches the team's convention per the `java-junit` skill).
   - Construct the exact minimal object graph / input that triggers `$EXC` at `$REL_PATH:$LINE_NO`.
   - Assert the current buggy behaviour using `assertThrows(<ExceptionClass>.class, ...)`.
3. Run the test inside the dev-env container and **confirm it fails with `$EXC`**:

```bash
docker exec dev-env-dev-1 bash -c \
  "cd /home/dev/repo/ws/$SERVICE && mvn -pl . test -Dtest=StateProcessorNullSafetyTest -q"
```

Expected: the test case throws `$EXC`. If the test passes instead, your reproduction is wrong — fix the test, not the code. Do not proceed to step 5 until the test is **confirmed red with the exact exception**.

Then **flip the assertion** to describe the desired post-fix behaviour (e.g. `assertDoesNotThrow`, or `assertEquals(expected, actual)`). The test is now red, but for the right reason — it asserts the fix works.

### 5. GREEN — implement the minimal fix

Edit `$REL_PATH` (or a closely-related file identified by the RCA) with the smallest change that makes the test pass. Prefer:
- Null guards / `Optional.ofNullable` wrapping for KTable lookups
- Defensive early-return on empty reporter state
- Single-line version pins in `pom.xml` for Protobuf/Avro drift

Do **not** refactor beyond the fix. Do **not** edit comments or unrelated tests.

Re-run the same command from step 4 and **confirm it passes**:

```bash
docker exec dev-env-dev-1 bash -c \
  "cd /home/dev/repo/ws/$SERVICE && mvn -pl . test -Dtest=StateProcessorNullSafetyTest -q"
```

If it still fails, iterate (edit → re-test). If it passes, continue.

### 6. Full quality gate — use the `ctl` skill

The `ctl` skill runs compile → unit tests → lint source code → lint k8s objects inside `dev-env-dev-1`. Invoke it exactly as the CVE flow does. **Do not substitute your own `mvn verify`** — `ctl` covers more checks and matches what Jenkins will run later.

If `ctl` fails:
- Post a comment on the current issue with the last 80 lines of the failed step's log
- Set the issue status to `blocked`
- Post a Slack alert by curling `$SLACK_WEBHOOK_URL` directly with the `validation_failed` template from the Verifier's `slack-templates.md` (shape is shared — Fixer uses the same webhook, different outcome string).
- Do **NOT** spawn a Verifier child. Exit.

### 7. Commit + push

```bash
git add -A
git commit -m "fix($SERVICE): resolve $EXC at $REL_PATH:$LINE_NO

Root cause: <one-sentence summary>.
TDD: failing test added as <test-class>, now green.
Validated via ctl quality gate.

${JIRA_KEY}"
git push -u origin "$BRANCH"
```

### 8. Spawn Verifier child

Use the `multica-cli` skill. `--assignee` takes the agent name directly — no UUIDs:

```bash
multica issue create \
  --title "$JIRA_KEY: Raise PR for $SERVICE $EXC fix" \
  --description "$(cat /tmp/verifier-payload.md)" \
  --assignee "IoTOps Verifier" \
  --parent "$FIXER_ISSUE_ID" \
  --priority high --status todo \
  --output json
```

Then `multica issue status "$FIXER_ISSUE_ID" done` on your own issue. Both commands documented in `multica-cli`'s `issue-ops.md`.

Verifier description (headings are parsed by the Verifier skill):

```markdown
## Fix Ready

**Service**: `<service>`
**Branch**: `<bugfix/…>`
**Jira**: <jira_url>
**Jira Key**: <JIRA_KEY>
**Humio**: <humio_link>
**Fixer Issue**: <multica_url_for_this_issue>

## TDD Summary

- Failing test added: `<test-class-name>` (red → green)
- Production fix: `<relative_path>:<line>`
- Quality gate: `ctl` PASS (compile + tests + source lint + k8s lint)

## Commit
```
<git log -1 --format="%B" output>
```

## Patch
```diff
<git diff devel..HEAD>
```
```

Clean up `WORKDIR`.

## Do not

- Do **not** open a PR or touch Jira status — that's the Verifier.
- Do **not** skip the RED step even when "the fix is obvious" — the failing test becomes the regression guard.
- Do **not** include refactors or formatting churn. Reviewers should see a one- to three-line production diff plus one new test file.
- Do **not** push a branch that failed `ctl`. Jenkins will re-run the same checks and reject it anyway.
