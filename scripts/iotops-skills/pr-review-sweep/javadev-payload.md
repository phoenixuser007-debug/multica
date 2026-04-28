# Java-Dev-Agent child-issue payload template

Reference template for SKILL.md Step 4. Render one of these per PR with unresolved review comments and pass it as the `--description` of the child issue (Step 5).

## Template

```markdown
## PR Review Action Required

**PR**: $PR_URL
**Service**: $REPO
**Branch**: $BRANCH
**Jira Key**: $JIRA_KEY
**Sweep Issue**: $THIS_ISSUE_URL
**Fixer Bot Author**: $FIXER_OR_BRIDGE_USER

## Review Comments

<one ### block per top-level thread, grouped by (file, line)>

### `$PATH:$LINE` — $AUTHOR_DISPLAY_NAME (@$AUTHOR_LOGIN)

> $COMMENT_TEXT_LINE_1
> $COMMENT_TEXT_LINE_2
> ...

<reply chain, indented two spaces under each parent, in order:>
  - **@$REPLY_AUTHOR_LOGIN**: $REPLY_TEXT
  - **@$REPLY_AUTHOR_LOGIN**: $REPLY_TEXT

<repeat ### block per thread>

## What you must do

1. Use the existing checkout at `/root/dev-env/ws/$REPO`. Do **not** clone to `/tmp`.
2. `git fetch origin $BRANCH && git checkout $BRANCH && git reset --hard origin/$BRANCH`.
3. Address each review thread above with the **smallest possible change**:
   - For each thread, decide: address (apply the change) or decline (with a one-line reason).
   - Prefer the reviewer's wording when it matches an existing pattern in the file.
   - Do **not** refactor beyond the comments. No formatting churn, no rename sweeps.
4. If new behaviour is introduced, add a JUnit test that locks it in (TDD: red → green). Match the convention from the original Fixer commit on this branch.
5. Run the `ctl` quality gate. It must pass.
6. Commit + push the **same branch** (no rename, no new branch — Bitbucket re-uses the PR):
   ```bash
   git add -A
   git commit -m "review($REPO): address PR #$PR_ID feedback

   Threads addressed: <N>. Threads declined: <M> (see PR replies for reasons).

   $JIRA_KEY"
   git push origin $BRANCH
   ```
7. On the PR itself, reply to **every** thread above via `pr-cli reply`:
   - **Addressed:** `Done in <short-sha>: <one-line summary of the change>.`
   - **Declined:** `Won't change because <one-line reason>.` (Use sparingly — when in doubt, address.)
8. When the push succeeds and replies are posted, close out:
   ```bash
   multica issue comment add "$THIS_CHILD_ISSUE_ID" --content-stdin <<EOF
   ## Done
   Pushed: $(git rev-parse HEAD)
   Threads addressed: <N>
   Threads declined: <M>
   PR: $PR_URL
   EOF
   multica issue status "$THIS_CHILD_ISSUE_ID" done
   ```
9. The sweep issue (`$THIS_ISSUE_URL`) and the original Fixer issue stay where they are — you do not transition them.

## Do not

- Do **not** rebase the branch onto `devel`. Bitbucket will recompute the PR diff and lose review-comment anchors. Plain commits on top.
- Do **not** force-push unless the only path is to amend a TDD test (then prefix the commit message `review!:` so reviewers see the rewrite).
- Do **not** auto-resolve threads in Bitbucket. The reviewer marks resolved, not you.
- Do **not** re-open or edit the PR description. The Verifier owns it.
- Do **not** spawn further multica children. This is a leaf node — the only outputs are the push + the PR replies + the `done` transition on this issue.

## If you can't address a comment

If a thread requires information you don't have (e.g. business-rule clarification, missing access to a schema), reply on the PR with `Needs reviewer clarification: <question>`, leave the thread open, and transition this issue to `blocked` with a comment quoting the unresolved threads. The autopilot will not re-spawn — a human will pick it up from `blocked`.
```

## Renderer notes (for the sweep skill)

When building the description from `/tmp/pr-${PR_ID}-review.json`:

- Group items by `(commentAnchor.path, commentAnchor.line)`. Stash returns one entry per top-level thread already; replies live under `.comments[]`.
- The body of `> blockquote` lines must be the comment's `text` field, prefixed `> ` per line. Don't escape Markdown inside it — reviewers often quote code, and re-escaping garbles it.
- For the reply chain, walk `.comments[]` recursively (it can nest); render two-space-indented bullets, one bullet per reply, in chronological order.
- If `commentAnchor.line` is null but `commentAnchor.path` is set, label the thread `### \`$PATH\` (file-level) — $AUTHOR`.
- Strip the trailing fixer-bot comments **before** grouping (already done by the Step 2 jq filter), so the rendered payload never includes status updates.
