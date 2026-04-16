import { useState, useCallback } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api";
import { issueKeys, CLOSED_PAGE_SIZE, type MyIssuesFilter } from "./queries";
import { useWorkspaceId } from "../hooks";
import type { Issue, IssueReaction } from "../types";
import type {
  CreateIssueRequest,
  UpdateIssueRequest,
  ListIssuesResponse,
} from "../types";
import type { TimelineEntry, IssueSubscriber, Reaction } from "../types";

// ---------------------------------------------------------------------------
// CVE Remediation workflow
// ---------------------------------------------------------------------------

export interface CVERemediationRepo {
  name: string;
  /** Path inside the dev container (e.g. /home/dev/repo/ws/<repo>) — for docker exec commands */
  path: string;
  /** Path on the host (e.g. /root/dev-env/ws/<repo>) — for git operations run directly */
  hostPath: string;
  branch: string;
  jiraSubtaskKey?: string;
  jiraStoryKey?: string;
  projectId?: string;
  parentIssueId?: string;
}

export interface CVERemediationProgress {
  total: number;
  completed: number;
  current: string | null;
  issueIds: string[];
  reusedIds: string[];
  errors: { repo: string; error: string }[];
}

export function buildCVEIssueDescription(repo: CVERemediationRepo): string {
  return [
    `CVE scan and remediation for \`${repo.name}\`.`,
    ``,
    `**Repo path (container):** \`${repo.path}\``,
    `**Repo path (host):** \`${repo.hostPath}\``,
    `**Branch:** \`${repo.branch}\``,
    ...(repo.jiraSubtaskKey ? [
      `**Jira:** [${repo.jiraSubtaskKey}](https://jira.arubanetworks.com/browse/${repo.jiraSubtaskKey})` +
      (repo.jiraStoryKey ? ` · Story: [${repo.jiraStoryKey}](https://jira.arubanetworks.com/browse/${repo.jiraStoryKey})` : ""),
    ] : []),
    ``,
    `## State transitions — you own these`,
    ``,
    `| When | Multica | Jira (${repo.jiraSubtaskKey ?? "sub-task key"}) |`,
    `|------|---------|------|`,
    `| You start reading this task | \`todo\` | — |`,
    `| You begin the trivy scan | \`in_progress\` | transition → **Assigned** |`,
    `| Trivy scan is clean (0 HIGH 0 CRITICAL) — skip PR | \`done\` | transition → **Resolved** |`,
    `| Before raising PR (CVEs found) | — | transition → **In Review** |`,
    `| PR is open | \`in_review\` | — |`,
    `| PR merged / ctl passes | \`done\` | — |`,
    `| User comments on this issue | back to \`todo\` → \`in_progress\` | — |`,
    ``,
    `## Step 1 — Read project build instructions`,
    ``,
    `Check if \`${repo.path}/AGENTS.md\` exists:`,
    `- **If it exists** — follow its instructions exactly for compile, test, and lint commands.`,
    `- **If it does not exist** — use the standard \`tools/\` scripts (all run inside the dev container):`,
    `\`\`\``,
    `docker exec -w /home/dev/repo/ws/${repo.name} dev-env-dev-1 bash tools/compile.sh`,
    `docker exec -w /home/dev/repo/ws/${repo.name} dev-env-dev-1 bash tools/unit-tests.sh`,
    `docker exec -w /home/dev/repo/ws/${repo.name} dev-env-dev-1 bash tools/lint-source-code.sh`,
    `docker exec -w /home/dev/repo/ws/${repo.name} dev-env-dev-1 bash tools/lint-k8s-objects.sh`,
    `\`\`\``,
    ``,
    `## Step 2 — Sync to latest devel and prepare branch`,
    ``,
    `> **Important:** The repo is already cloned at \`${repo.hostPath}\` on the host. **Do NOT use \`multica repo checkout\`** — just use the path directly.`,
    ``,
    `Move this Multica issue to \`in_progress\` and transition the Jira sub-task to **Assigned**:`,
    `\`\`\``,
    ...(repo.jiraSubtaskKey
      ? [`python3 /root/.claude/skills/jira-cli/scripts/jira.py transition ${repo.jiraSubtaskKey} "Assign"`]
      : [`# python3 /root/.claude/skills/jira-cli/scripts/jira.py transition <JIRA-KEY> "Assign"`]),
    `\`\`\``,
    ``,
    `Sync to latest devel — run these **every time** (on the host):`,
    `\`\`\``,
    `cd ${repo.hostPath}`,
    `git checkout devel`,
    `git pull origin devel`,
    `git checkout -B ${repo.branch}`,
    `\`\`\``,
    `\`git checkout -B\` force-recreates the branch from the freshly pulled devel, discarding any previous scan attempt on that branch.`,
    ``,
    `## Step 3 — Scan and fix CVEs`,
    ``,
    `Run the **trivy-cve-remediation** skill. It will scan the built JARs and upgrade vulnerable dependency versions in \`pom.xml\`.`,
    ``,
    `**If trivy reports 0 HIGH and 0 CRITICAL** — the repo is already clean. Skip Steps 4–5 and go directly to Step 6 (already-clean path).`,
    ``,
    `## Step 4 — Quality gate (ctl)`,
    ``,
    `_(Skip this step if trivy was already clean in Step 3.)_`,
    ``,
    `Run the **ctl** skill to compile, test, and lint the repo using the build commands from Step 1.`,
    ``,
    `**If ctl fails:**`,
    `- Read the error output carefully.`,
    `- If a version bump broke an API: find the nearest compatible fixed version and update \`pom.xml\`.`,
    `- If tests fail: fix the code or the test to match the new dependency behaviour.`,
    `- Re-run ctl. Repeat until it passes. Do not move to \`in_review\` until ctl is green.`,
    ``,
    `## Step 5 — Raise PR`,
    ``,
    `_(Skip this step if trivy was already clean in Step 3 — no code changes were made, no PR needed.)_`,
    ``,
    `Once ctl is green, first transition Jira to **In Review**, then push and create the PR:`,
    `\`\`\``,
    ...(repo.jiraSubtaskKey
      ? [`python3 /root/.claude/skills/jira-cli/scripts/jira.py transition ${repo.jiraSubtaskKey} "In Review"`]
      : [`# python3 /root/.claude/skills/jira-cli/scripts/jira.py transition <JIRA-KEY> "In Review"`]),
    `git push origin ${repo.branch} --force-with-lease`,
    ...(repo.jiraSubtaskKey
      ? [
          `pr-cli create --title "${repo.jiraSubtaskKey}: chore: CVE remediation ${repo.name}" --description "Jira: ${repo.jiraSubtaskKey}\\n\\nAutomated CVE remediation via Trivy scan. ctl quality gate passed." --target devel`,
        ]
      : [
          `pr-cli create --title "chore: CVE remediation ${repo.name}" --description "Automated CVE remediation via Trivy scan. ctl quality gate passed." --target devel`,
        ]),
    `\`\`\``,
    `(Use \`--force-with-lease\` because the branch is force-recreated from devel each run.)`,
    `Move Multica issue to \`in_review\`.`,
    ``,
    `## Step 6 — Done`,
    ``,
    `**Already-clean path** (trivy was 0→0, no PR raised):`,
    `\`\`\``,
    ...(repo.jiraSubtaskKey
      ? [`python3 /root/.claude/skills/jira-cli/scripts/jira.py transition ${repo.jiraSubtaskKey} "Resolve Issue"`]
      : [`# python3 /root/.claude/skills/jira-cli/scripts/jira.py transition <JIRA-KEY> "Resolve Issue"`]),
    `\`\`\``,
    ``,
    `Move this Multica issue to \`done\`, then notify Slack:`,
    `\`\`\``,
    `curl -s -X POST http://localhost:8090/api/slack/cve-done \\`,
    `  -H "Content-Type: application/json" \\`,
    `  -H "X-Workspace-ID: $MULTICA_WORKSPACE_ID" \\`,
    `  -H "Authorization: Bearer $MULTICA_TOKEN" \\`,
    `  -d '{`,
    `    "repo": "${repo.name}",`,
    ...(repo.jiraSubtaskKey
      ? [
          `    "jira_key": "${repo.jiraSubtaskKey}",`,
          `    "jira_url": "https://jira.arubanetworks.com/browse/${repo.jiraSubtaskKey}",`,
        ]
      : []),
    `    "pr_title": "<PR title, or \\"Already clean — no PR needed\\" if trivy was 0→0>",`,
    `    "pr_url": "<PR URL, or \\"\\" if no PR was raised>",`,
    `    "cve_high_before": <HIGH count from trivy scan>,`,
    `    "cve_critical_before": <CRITICAL count from trivy scan>,`,
    `    "cve_high_after": <HIGH count after remediation — same as before if already clean>,`,
    `    "cve_critical_after": <CRITICAL count after remediation — same as before if already clean>`,
    `  }'`,
    `\`\`\``,
    `For CVE counts: parse the trivy scan output — count occurrences of "HIGH" and "CRITICAL" severity lines. If trivy was already clean (0→0), set both before and after counts to 0.`,
    ``,
    `## If the user comments`,
    ``,
    `When you receive a new comment on this issue at any point:`,
    `1. Move issue to \`todo\` to acknowledge.`,
    `2. Move to \`in_progress\` when you start addressing it.`,
    `3. Read the comment, apply any changes requested, re-run ctl, update the PR.`,
    `4. Move back to \`in_review\` (or \`done\` if fully resolved).`,
  ].join("\n");
}

export function useCVERemediation() {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  const [progress, setProgress] = useState<CVERemediationProgress>({
    total: 0,
    completed: 0,
    current: null,
    issueIds: [],
    reusedIds: [],
    errors: [],
  });
  const [isRunning, setIsRunning] = useState(false);

  const run = useCallback(
    async (repos: CVERemediationRepo[], agentId: string) => {
      setIsRunning(true);
      setProgress({ total: repos.length, completed: 0, current: null, issueIds: [], reusedIds: [], errors: [] });

      const issueIds: string[] = [];
      const reusedIds: string[] = [];
      const errors: { repo: string; error: string }[] = [];

      // Fetch all open CVE remediation issues once — avoids N individual searches
      const monthLabel = new Date().toISOString().slice(0, 7); // "2026-04"
      // Key: repo name (without Jira prefix) → { id, title }
      let openIssuesByRepo = new Map<string, { id: string; title: string }>();
      try {
        const searchResp = await api.searchIssues({
          q: `CVE Remediation`,
          limit: 500,
          include_closed: false,
        });
        for (const issue of searchResp.issues) {
          if (issue.created_at.startsWith(monthLabel)) {
            // Title may be "CVE Remediation: <repo>" or "CNX-XXXXX: CVE Remediation: <repo>"
            // Extract repo name from the end of the title after the last "CVE Remediation: " token
            const match = issue.title.match(/CVE Remediation:\s*(.+)$/);
            const repoName = match?.[1]?.trim();
            if (repoName) {
              openIssuesByRepo.set(repoName, { id: issue.id, title: issue.title });
            }
          }
        }
      } catch {
        // Non-fatal: fall through to always-create
      }

      for (const repo of repos) {
        setProgress((p) => ({ ...p, current: repo.name }));

        try {
          // Title leads with Jira key when available so it's the primary visible identifier
          const title = repo.jiraSubtaskKey
            ? `${repo.jiraSubtaskKey}: CVE Remediation: ${repo.name}`
            : `CVE Remediation: ${repo.name}`;

          const existing = openIssuesByRepo.get(repo.name);

          if (existing) {
            // Reuse the existing open ticket — refresh title (may need Jira key added), description, assignee
            const desc = buildCVEIssueDescription(repo);
            await api.updateIssue(existing.id, {
              title,
              description: desc,
              assignee_type: "agent",
              assignee_id: agentId,
              ...(repo.projectId ? { project_id: repo.projectId } : {}),
              ...(repo.parentIssueId ? { parent_issue_id: repo.parentIssueId } : {}),
            });
            issueIds.push(existing.id);
            reusedIds.push(existing.id);
          } else {
            // Create a fresh ticket at backlog with agent assigned — agent drives all transitions
            const desc = buildCVEIssueDescription(repo);
            const issue = await api.createIssue({
              title,
              description: desc,
              status: "backlog",
              assignee_type: "agent",
              assignee_id: agentId,
              ...(repo.projectId ? { project_id: repo.projectId } : {}),
              ...(repo.parentIssueId ? { parent_issue_id: repo.parentIssueId } : {}),
            });
            issueIds.push(issue.id);
          }
        } catch (err) {
          errors.push({ repo: repo.name, error: String(err) });
        }

        setProgress((p) => ({
          ...p,
          completed: p.completed + 1,
          issueIds,
          reusedIds,
          errors,
        }));
      }

      setProgress((p) => ({ ...p, current: null, issueIds, reusedIds, errors }));
      setIsRunning(false);
      qc.invalidateQueries({ queryKey: issueKeys.list(wsId) });

      return { issueIds, reusedIds, errors };
    },
    [qc, wsId],
  );

  return { run, isRunning, progress };
}

// ---------------------------------------------------------------------------
// Shared mutation variable types — used by both mutation hooks and
// useMutationState consumers to keep the type assertion in sync.
// ---------------------------------------------------------------------------

export type ToggleCommentReactionVars = {
  commentId: string;
  emoji: string;
  existing: Reaction | undefined;
};

export type ToggleIssueReactionVars = {
  emoji: string;
  existing: IssueReaction | undefined;
};

// ---------------------------------------------------------------------------
// Done issue pagination
// ---------------------------------------------------------------------------

export function useLoadMoreDoneIssues(myIssues?: { scope: string; filter: MyIssuesFilter }) {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  const [isLoading, setIsLoading] = useState(false);

  const queryKey = myIssues
    ? issueKeys.myList(wsId, myIssues.scope, myIssues.filter)
    : issueKeys.list(wsId);
  const cache = qc.getQueryData<ListIssuesResponse>(queryKey);
  const doneLoaded = cache
    ? cache.issues.filter((i) => i.status === "done").length
    : 0;
  const doneTotal = cache?.doneTotal ?? 0;
  const hasMore = doneLoaded < doneTotal;

  const loadMore = useCallback(async () => {
    if (isLoading || !hasMore) return;
    setIsLoading(true);
    try {
      const res = await api.listIssues({
        status: "done",
        limit: CLOSED_PAGE_SIZE,
        offset: doneLoaded,
        ...myIssues?.filter,
      });
      qc.setQueryData<ListIssuesResponse>(queryKey, (old) => {
        if (!old) return old;
        const existingIds = new Set(old.issues.map((i) => i.id));
        const newIssues = res.issues.filter((i) => !existingIds.has(i.id));
        return {
          ...old,
          issues: [...old.issues, ...newIssues],
          doneTotal: res.total,
        };
      });
    } finally {
      setIsLoading(false);
    }
  }, [qc, queryKey, doneLoaded, hasMore, isLoading, myIssues?.filter]);

  return { loadMore, hasMore, isLoading, doneTotal };
}

// ---------------------------------------------------------------------------
// Issue CRUD
// ---------------------------------------------------------------------------

export function useCreateIssue() {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: (data: CreateIssueRequest) => api.createIssue(data),
    onSuccess: (newIssue) => {
      qc.setQueryData<ListIssuesResponse>(issueKeys.list(wsId), (old) =>
        old && !old.issues.some((i) => i.id === newIssue.id)
          ? {
              ...old,
              issues: [...old.issues, newIssue],
              total: old.total + 1,
              doneTotal: (old.doneTotal ?? 0) + (newIssue.status === "done" ? 1 : 0),
            }
          : old,
      );
      // Invalidate parent's children query so sub-issues list updates immediately
      if (newIssue.parent_issue_id) {
        qc.invalidateQueries({ queryKey: issueKeys.children(wsId, newIssue.parent_issue_id) });
        qc.invalidateQueries({ queryKey: issueKeys.childProgress(wsId) });
      }
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: issueKeys.list(wsId) });
    },
  });
}

export function useUpdateIssue() {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string } & UpdateIssueRequest) =>
      api.updateIssue(id, data),
    onMutate: ({ id, ...data }) => {
      // Fire-and-forget cancelQueries — keeps onMutate synchronous so the
      // cache update happens in the same tick as mutate(). Awaiting would
      // yield to the event loop, letting @dnd-kit reset its visual state
      // before the optimistic update lands.
      qc.cancelQueries({ queryKey: issueKeys.list(wsId) });
      const prevList = qc.getQueryData<ListIssuesResponse>(issueKeys.list(wsId));
      const prevDetail = qc.getQueryData<Issue>(issueKeys.detail(wsId, id));

      // Resolve parent_issue_id from the freshest source so we can keep the
      // parent's children cache in sync (used by the parent issue's
      // sub-issues list).
      const parentId =
        prevDetail?.parent_issue_id ??
        prevList?.issues.find((i) => i.id === id)?.parent_issue_id ??
        null;
      const prevChildren = parentId
        ? qc.getQueryData<Issue[]>(issueKeys.children(wsId, parentId))
        : undefined;

      qc.setQueryData<ListIssuesResponse>(issueKeys.list(wsId), (old) =>
        old
          ? {
              ...old,
              issues: old.issues.map((i) =>
                i.id === id ? { ...i, ...data } : i,
              ),
            }
          : old,
      );
      qc.setQueryData<Issue>(issueKeys.detail(wsId, id), (old) =>
        old ? { ...old, ...data } : old,
      );
      if (parentId) {
        qc.setQueryData<Issue[]>(
          issueKeys.children(wsId, parentId),
          (old) =>
            old?.map((c) => (c.id === id ? { ...c, ...data } : c)),
        );
      }
      return { prevList, prevDetail, prevChildren, parentId, id };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.prevList) qc.setQueryData(issueKeys.list(wsId), ctx.prevList);
      if (ctx?.prevDetail)
        qc.setQueryData(issueKeys.detail(wsId, ctx.id), ctx.prevDetail);
      if (ctx?.parentId && ctx.prevChildren !== undefined) {
        qc.setQueryData(
          issueKeys.children(wsId, ctx.parentId),
          ctx.prevChildren,
        );
      }
    },
    onSettled: (_data, _err, vars, ctx) => {
      qc.invalidateQueries({ queryKey: issueKeys.detail(wsId, vars.id) });
      qc.invalidateQueries({ queryKey: issueKeys.list(wsId) });
      // Invalidate old parent's children cache
      if (ctx?.parentId) {
        qc.invalidateQueries({
          queryKey: issueKeys.children(wsId, ctx.parentId),
        });
        qc.invalidateQueries({ queryKey: issueKeys.childProgress(wsId) });
      }
      // Invalidate new parent's children cache when parent_issue_id changed
      const newParentId = vars.parent_issue_id;
      if (newParentId && newParentId !== ctx?.parentId) {
        qc.invalidateQueries({
          queryKey: issueKeys.children(wsId, newParentId),
        });
        qc.invalidateQueries({ queryKey: issueKeys.childProgress(wsId) });
      }
    },
  });
}

export function useDeleteIssue() {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: (id: string) => api.deleteIssue(id),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: issueKeys.list(wsId) });
      const prevList = qc.getQueryData<ListIssuesResponse>(issueKeys.list(wsId));
      const deleted = prevList?.issues.find((i) => i.id === id);
      qc.setQueryData<ListIssuesResponse>(issueKeys.list(wsId), (old) => {
        if (!old) return old;
        const d = old.issues.find((i) => i.id === id);
        return {
          ...old,
          issues: old.issues.filter((i) => i.id !== id),
          total: old.total - 1,
          doneTotal: (old.doneTotal ?? 0) - (d?.status === "done" ? 1 : 0),
        };
      });
      qc.removeQueries({ queryKey: issueKeys.detail(wsId, id) });
      return { prevList, parentIssueId: deleted?.parent_issue_id };
    },
    onError: (_err, _id, ctx) => {
      if (ctx?.prevList) qc.setQueryData(issueKeys.list(wsId), ctx.prevList);
    },
    onSettled: (_data, _err, _id, ctx) => {
      qc.invalidateQueries({ queryKey: issueKeys.list(wsId) });
      if (ctx?.parentIssueId) {
        qc.invalidateQueries({ queryKey: issueKeys.children(wsId, ctx.parentIssueId) });
        qc.invalidateQueries({ queryKey: issueKeys.childProgress(wsId) });
      }
    },
  });
}

export function useBatchUpdateIssues() {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: ({
      ids,
      updates,
    }: {
      ids: string[];
      updates: UpdateIssueRequest;
    }) => api.batchUpdateIssues(ids, updates),
    onMutate: async ({ ids, updates }) => {
      await qc.cancelQueries({ queryKey: issueKeys.list(wsId) });
      const prevList = qc.getQueryData<ListIssuesResponse>(issueKeys.list(wsId));
      qc.setQueryData<ListIssuesResponse>(issueKeys.list(wsId), (old) =>
        old
          ? {
              ...old,
              issues: old.issues.map((i) =>
                ids.includes(i.id) ? { ...i, ...updates } : i,
              ),
            }
          : old,
      );
      return { prevList };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.prevList) qc.setQueryData(issueKeys.list(wsId), ctx.prevList);
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: issueKeys.list(wsId) });
    },
  });
}

export function useBatchDeleteIssues() {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: (ids: string[]) => api.batchDeleteIssues(ids),
    onMutate: async (ids) => {
      await qc.cancelQueries({ queryKey: issueKeys.list(wsId) });
      const prevList = qc.getQueryData<ListIssuesResponse>(issueKeys.list(wsId));
      const idSet = new Set(ids);
      const parentIssueIds = new Set(
        prevList?.issues
          .filter((i) => idSet.has(i.id) && i.parent_issue_id)
          .map((i) => i.parent_issue_id!) ?? [],
      );
      qc.setQueryData<ListIssuesResponse>(issueKeys.list(wsId), (old) => {
        if (!old) return old;
        const doneDeleted = old.issues.filter(
          (i) => idSet.has(i.id) && i.status === "done",
        ).length;
        return {
          ...old,
          issues: old.issues.filter((i) => !idSet.has(i.id)),
          total: old.total - ids.length,
          doneTotal: (old.doneTotal ?? 0) - doneDeleted,
        };
      });
      return { prevList, parentIssueIds };
    },
    onError: (_err, _ids, ctx) => {
      if (ctx?.prevList) qc.setQueryData(issueKeys.list(wsId), ctx.prevList);
    },
    onSettled: (_data, _err, _ids, ctx) => {
      qc.invalidateQueries({ queryKey: issueKeys.list(wsId) });
      if (ctx?.parentIssueIds && ctx.parentIssueIds.size > 0) {
        for (const parentId of ctx.parentIssueIds) {
          qc.invalidateQueries({ queryKey: issueKeys.children(wsId, parentId) });
        }
        qc.invalidateQueries({ queryKey: issueKeys.childProgress(wsId) });
      }
    },
  });
}

// ---------------------------------------------------------------------------
// Comments / Timeline
// ---------------------------------------------------------------------------

export function useCreateComment(issueId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      content,
      type,
      parentId,
      attachmentIds,
    }: {
      content: string;
      type?: string;
      parentId?: string;
      attachmentIds?: string[];
    }) => api.createComment(issueId, content, type, parentId, attachmentIds),
    onSuccess: (comment) => {
      qc.setQueryData<TimelineEntry[]>(
        issueKeys.timeline(issueId),
        (old) => {
          if (!old) return old;
          const entry: TimelineEntry = {
            type: "comment",
            id: comment.id,
            actor_type: comment.author_type,
            actor_id: comment.author_id,
            content: comment.content,
            parent_id: comment.parent_id,
            comment_type: comment.type,
            reactions: comment.reactions ?? [],
            attachments: comment.attachments ?? [],
            created_at: comment.created_at,
            updated_at: comment.updated_at,
          };
          if (old.some((e) => e.id === comment.id)) return old;
          return [...old, entry];
        },
      );
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: issueKeys.timeline(issueId) });
    },
  });
}

export function useUpdateComment(issueId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ commentId, content }: { commentId: string; content: string }) =>
      api.updateComment(commentId, content),
    onMutate: async ({ commentId, content }) => {
      await qc.cancelQueries({ queryKey: issueKeys.timeline(issueId) });
      const prev = qc.getQueryData<TimelineEntry[]>(issueKeys.timeline(issueId));
      qc.setQueryData<TimelineEntry[]>(
        issueKeys.timeline(issueId),
        (old) =>
          old?.map((e) => (e.id === commentId ? { ...e, content } : e)),
      );
      return { prev };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.prev)
        qc.setQueryData(issueKeys.timeline(issueId), ctx.prev);
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: issueKeys.timeline(issueId) });
    },
  });
}

export function useDeleteComment(issueId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (commentId: string) => api.deleteComment(commentId),
    onMutate: async (commentId) => {
      await qc.cancelQueries({ queryKey: issueKeys.timeline(issueId) });
      const prev = qc.getQueryData<TimelineEntry[]>(issueKeys.timeline(issueId));

      // Cascade: collect all child comment IDs
      const toRemove = new Set<string>([commentId]);
      if (prev) {
        let changed = true;
        while (changed) {
          changed = false;
          for (const e of prev) {
            if (e.parent_id && toRemove.has(e.parent_id) && !toRemove.has(e.id)) {
              toRemove.add(e.id);
              changed = true;
            }
          }
        }
      }

      qc.setQueryData<TimelineEntry[]>(
        issueKeys.timeline(issueId),
        (old) => old?.filter((e) => !toRemove.has(e.id)),
      );
      return { prev };
    },
    onError: (_err, _id, ctx) => {
      if (ctx?.prev)
        qc.setQueryData(issueKeys.timeline(issueId), ctx.prev);
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: issueKeys.timeline(issueId) });
    },
  });
}

export function useToggleCommentReaction(issueId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationKey: ["toggleCommentReaction", issueId] as const,
    mutationFn: async ({
      commentId,
      emoji,
      existing,
    }: ToggleCommentReactionVars) => {
      if (existing) {
        await api.removeReaction(commentId, emoji);
        return null;
      }
      return api.addReaction(commentId, emoji);
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: issueKeys.timeline(issueId) });
    },
  });
}

// ---------------------------------------------------------------------------
// Issue-level Reactions
// ---------------------------------------------------------------------------

export function useToggleIssueReaction(issueId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationKey: ["toggleIssueReaction", issueId] as const,
    mutationFn: async ({
      emoji,
      existing,
    }: ToggleIssueReactionVars) => {
      if (existing) {
        await api.removeIssueReaction(issueId, emoji);
        return null;
      }
      return api.addIssueReaction(issueId, emoji);
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: issueKeys.reactions(issueId) });
    },
  });
}

// ---------------------------------------------------------------------------
// Issue Subscribers
// ---------------------------------------------------------------------------

export function useToggleIssueSubscriber(issueId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      userId,
      userType,
      subscribed,
    }: {
      userId: string;
      userType: "member" | "agent";
      subscribed: boolean;
    }) => {
      if (subscribed) {
        await api.unsubscribeFromIssue(issueId, userId, userType);
      } else {
        await api.subscribeToIssue(issueId, userId, userType);
      }
    },
    onMutate: async ({ userId, userType, subscribed }) => {
      await qc.cancelQueries({ queryKey: issueKeys.subscribers(issueId) });
      const prev = qc.getQueryData<IssueSubscriber[]>(
        issueKeys.subscribers(issueId),
      );

      if (subscribed) {
        qc.setQueryData<IssueSubscriber[]>(
          issueKeys.subscribers(issueId),
          (old) =>
            old?.filter(
              (s) => !(s.user_id === userId && s.user_type === userType),
            ),
        );
      } else {
        const temp: IssueSubscriber = {
          issue_id: issueId,
          user_type: userType,
          user_id: userId,
          reason: "manual",
          created_at: new Date().toISOString(),
        };
        qc.setQueryData<IssueSubscriber[]>(
          issueKeys.subscribers(issueId),
          (old) => {
            if (
              old?.some(
                (s) => s.user_id === userId && s.user_type === userType,
              )
            )
              return old;
            return [...(old ?? []), temp];
          },
        );
      }
      return { prev };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.prev)
        qc.setQueryData(issueKeys.subscribers(issueId), ctx.prev);
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: issueKeys.subscribers(issueId) });
    },
  });
}
