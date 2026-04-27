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
- `SLACK_BOT_TOKEN` + `SLACK_CHANNEL_ID` — for threaded Slack replies (optional; graceful no-op if absent)
- `SLACK_WEBHOOK_URL` — fallback for non-threaded posts when bot token is unset
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

After posting the comment, reply to the Scout's Slack thread with the RCA summary
so the team can track progress in real time:

```bash
# Load helpers from the shared skill (materialised alongside this skill at runtime)
source ~/.claude/skills/iotops-verify-pr-slack/slack-templates.md 2>/dev/null || true

# thread_ts was stored by Scout in a comment on this Fixer issue
SLACK_THREAD_TS=$(extract_thread_ts "$FIXER_ISSUE_ID")

RCA_MSG=$(printf ':mag: *RCA identified*\n*Service:* `%s`  |  *Exception:* `%s`\n*Root cause:* %s\n*Approach:* %s\nStarting TDD fix (red→green)…' \
  "$SERVICE" "$EXC" "$ROOT_CAUSE" "$APPROACH")

slack_post "$RCA_MSG" "$SLACK_THREAD_TS"
```

### 3. Use the pre-cloned repo and branch

The repos are already cloned at `/root/dev-env/ws/<service>` — the same path used by the CVE flow. **Do not clone to /tmp.** Work directly in the existing checkout:

```bash
REPO="/root/dev-env/ws/$SERVICE"
cd "$REPO"

# Sync to latest devel
git fetch origin devel
git checkout devel
git reset --hard origin/devel

SLUG=$(echo "$EXC" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | cut -c1-40)
BRANCH="bugfix/${JIRA_KEY}-${SLUG}"
git checkout -b "$BRANCH"
```

The container path for `docker exec` commands is `/home/dev/repo/ws/$SERVICE` (same directory, mounted into `dev-env-dev-1`).

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

### 5b. Coverage sanity check — the test must actually run

A passing focused test (`-Dtest=YourTest`) does **not** prove the test will run during the JaCoCo gate. If the test is in the wrong module, excluded by surefire patterns, missing from `target/test-classes/`, or skipped at runtime, JaCoCo will report `instructions covered ratio is 0.00` and the gate will block the push — wasting a full ctl run. Catch this BEFORE invoking ctl.

For the module that contains the change (call it `$MODULE`, e.g. `api`, `integration-test`, or `core`):

1. **Run the module's full test suite** (no `-Dtest=` filter), matching what JaCoCo will see during ctl:
   ```bash
   docker exec dev-env-dev-1 bash -c \
     "cd /home/dev/repo/ws/$SERVICE && mvn -pl $MODULE test -q"
   ```

2. **Verify the new test class compiled into the right module:**
   ```bash
   docker exec dev-env-dev-1 bash -c \
     "find /home/dev/repo/ws/$SERVICE/$MODULE/target/test-classes -name '<YourTest>.class'"
   ```
   If empty, the test is in the wrong module — move the source file under `$MODULE/src/test/java/...` and rerun.

3. **Verify the new test actually executed:**
   ```bash
   docker exec dev-env-dev-1 bash -c \
     "ls /home/dev/repo/ws/$SERVICE/$MODULE/target/surefire-reports/TEST-*<YourTest>*.xml"
   ```
   Read the XML and confirm `tests > 0` and `errors=0 failures=0`. If skipped, surefire excluded it — check `<surefire>` config in the parent `pom.xml`.

4. **Verify JaCoCo recorded coverage:**
   ```bash
   docker exec dev-env-dev-1 bash -c \
     "ls -la /home/dev/repo/ws/$SERVICE/$MODULE/target/jacoco.exec"
   ```
   The file must exist and be > 0 bytes. An empty / missing exec means the JaCoCo agent wasn't loaded — likely because the `mvn test` invocation needs `-Pcoverage` or similar profile that the project's parent pom defines.

5. **If JaCoCo bundle ratio is < threshold for the module that contains your change**, the test isn't pulling its weight (or isn't running). Fixing options before reaching ctl:
   - Add another test case to the same test class that exercises a different code path of the changed file (raises both instruction + branch coverage cheaply).
   - Move the test from a sub-module that has its own JaCoCo bundle into the module that owns the changed source.
   - If the project legitimately had pre-existing 0% coverage on that module (rare), this is a baseline-broken case — handle in step 6a.

Only proceed to step 6 (full ctl gate) once steps 5b.1–5b.4 all pass. **Do not declare the fix done while the new test isn't visible to the JaCoCo gate.** A focused test alone isn't a complete fix.

### 5a. Ensure git safe.directory inside the dev container

`tools/lint-source-code.sh` triggers a `pre-commit` hook that calls `git`
inside `dev-env-dev-1`. If git rejects the mounted repo with
`fatal: detected dubious ownership in repository at ...`, the lint fails
before any check runs. Set the wildcard once (idempotent — safe to run on
every Fixer pickup):

```bash
docker exec dev-env-dev-1 bash -c \
  "git config --global --get-all safe.directory | grep -qx '*' \
   || git config --global --add safe.directory '*'"
```

The setting is per-user in the container's home directory, so it survives
container restarts but **not** container recreation. Run this guard before
invoking `ctl`.

### 6. Full quality gate — use the `ctl` skill

The `ctl` skill runs compile → unit tests → lint source code → lint k8s objects inside `dev-env-dev-1`. Invoke it exactly as the CVE flow does. **Do not substitute your own `mvn verify`** — `ctl` covers more checks and matches what Jenkins will run later.

If `ctl` fails, you must distinguish between two cases — **don't auto-block on a baseline that's already broken on `devel`**:

#### 6a. Baseline check (only run on `ctl` failure)

Many bridge-5g / iotops repos currently have flaky integration tests, missing private-proto deps, or pre-existing JaCoCo/lint violations on `devel`. Blocking the fix in those cases hides our work behind unrelated infrastructure breakage. Re-run `ctl` against a **clean `devel` checkout** (no fix applied) and compare:

```bash
git stash push -m "fixer-baseline-check" --include-untracked
git checkout devel
git reset --hard origin/devel
# re-run the same ctl invocation that just failed
ctl_BASELINE_OUTPUT=$(... same ctl command ... 2>&1) || true
git checkout "$BRANCH"
git stash pop
```

Compare `ctl_BASELINE_OUTPUT` to the original failure tail. Two paths:

**Baseline ALSO fails on the same step (compile / unit-test / lint / coverage)** → the breakage is pre-existing, not caused by this fix. Do this:
1. Push the bugfix branch as **draft** anyway — Jenkins will reject it but the branch + commit history are preserved for review.
2. Post a comment on the current issue stating: "ctl failure reproduces on clean `origin/devel` (HEAD: `<sha>`) — pre-existing breakage. Pushed as draft for review."
3. Create a child INFRA issue assigned to the human owner (`--assignee` based on the repo's `CODEOWNERS` or fall back to `naveen.u.holla@hpe.com`):
   ```bash
   multica issue create \
     --title "INFRA: $SERVICE devel baseline broken — blocks Fixer auto-push" \
     --description "$(cat /tmp/baseline-report.md)" \
     --priority high \
     --parent-id "$FIXER_ISSUE_ID"
   ```
4. Set this issue status to `in_review` (NOT `blocked`) and **proceed to Step 8 to spawn the Verifier** — they'll pick up the draft PR.

**Baseline PASSES (failure introduced by this fix)** → fall through to 6b.

#### 6b. True fix-introduced failure

- Post a comment on the current issue with the last 80 lines of the failed step's log
- Set the issue status to `blocked`
- Post a Slack alert using the `validation_failed` template from `slack-templates.md`, threaded into the existing Scout thread:

```bash
source ~/.claude/skills/iotops-verify-pr-slack/slack-templates.md 2>/dev/null || true
SLACK_THREAD_TS=$(extract_thread_ts "$FIXER_ISSUE_ID")
VAL_MSG=$(printf ':warning: *Auto-fix FAILED compile/lint in dev-env*\n*Service:* `%s`  |  *Jira:* <%s|%s>\nNo PR raised — issue is blocked and needs human review.' \
  "$SERVICE" "$JIRA_URL" "$JIRA_KEY")
slack_post "$VAL_MSG" "$SLACK_THREAD_TS"
```

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
