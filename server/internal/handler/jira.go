package handler

// Aruba CVE remediation Jira flow.
//
// All Aruba-specific logic (field IDs, project keys, component map, sprint
// format quirk, JQL templates, assignee, fix-version naming) lives in the
// `cve-jira` skill at scripts/cve-skills/cve-jira/cve_jira.py. This handler
// is pure orchestration: parse the request, plan the work, fan out parallel
// shell-outs, assemble the response. To change Aruba quirks, edit the script
// — no Go rebuild required.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// --- API contract ----------------------------------------------------------

// JiraTicket is a single Jira issue surfaced in the response.
type JiraTicket struct {
	Key      string `json:"key"`
	URL      string `json:"url"`
	Existing bool   `json:"existing"`
}

// JiraRepoTicket pairs a repo with its sub-task ticket and parent story.
type JiraRepoTicket struct {
	Repo       string `json:"repo"`
	Component  string `json:"component"`
	StoryKey   string `json:"story_key"`
	SubtaskKey string `json:"subtask_key"`
	SubtaskURL string `json:"subtask_url"`
	Existing   bool   `json:"existing"`
}

// CreateCVEJiraTicketsRequest lists the repos to create Jira tickets for.
type CreateCVEJiraTicketsRequest struct {
	Repos []string `json:"repos"`
}

// CreateCVEJiraTicketsResponse returns the mapping of repo → jira keys.
type CreateCVEJiraTicketsResponse struct {
	Stories map[string]JiraTicket `json:"stories"`
	Repos   []JiraRepoTicket      `json:"repos"`
	Errors  map[string]string     `json:"errors"`
}

// --- Skill invocation ------------------------------------------------------

// Container path; in dev we fall back to the repo-relative path so `go run`
// works without env tweaks. Override with $CVE_JIRA_SCRIPT for tests.
const cveJiraScriptContainer = "/app/scripts/cve-skills/cve-jira/cve_jira.py"
const cveJiraScriptDev = "scripts/cve-skills/cve-jira/cve_jira.py"

func cveJiraScriptPath() string {
	if p := os.Getenv("CVE_JIRA_SCRIPT"); p != "" {
		return p
	}
	if _, err := os.Stat(cveJiraScriptContainer); err == nil {
		return cveJiraScriptContainer
	}
	return cveJiraScriptDev
}

// runCveJira execs the skill and returns its stdout JSON.
func runCveJira(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "python3", append([]string{cveJiraScriptPath()}, args...)...)
	cmd.Env = os.Environ() // inherit JIRA_TOKEN
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("cve_jira.py %s: %s", args[0], strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, err
	}
	return out, nil
}

type cveJiraTicketOut struct {
	Key      string `json:"key"`
	URL      string `json:"url"`
	Existing bool   `json:"existing"`
	Error    string `json:"error,omitempty"`
}

// callTicket runs a find-or-create-* command and parses the resulting ticket.
func callTicket(ctx context.Context, args ...string) (cveJiraTicketOut, error) {
	out, err := runCveJira(ctx, args...)
	if err != nil {
		return cveJiraTicketOut{}, err
	}
	var t cveJiraTicketOut
	if err := json.Unmarshal(out, &t); err != nil {
		return cveJiraTicketOut{}, fmt.Errorf("parse: %w (raw: %s)", err, out)
	}
	if t.Error != "" {
		return cveJiraTicketOut{}, errors.New(t.Error)
	}
	if t.Key == "" {
		return cveJiraTicketOut{}, fmt.Errorf("no key in response: %s", out)
	}
	return t, nil
}

type cveJiraPlan struct {
	Groups []cveJiraPlanGroup `json:"groups"`
}

type cveJiraPlanGroup struct {
	Group     string   `json:"group"`
	Project   string   `json:"project"`
	Component string   `json:"component"`
	Repos     []string `json:"repos"`
}

func planRepos(ctx context.Context, repos []string) (*cveJiraPlan, error) {
	f, err := os.CreateTemp("", "cve-repos-*.txt")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	for _, r := range repos {
		if _, err := fmt.Fprintln(f, r); err != nil {
			f.Close()
			return nil, err
		}
	}
	f.Close()

	out, err := runCveJira(ctx, "plan", "--repos-file", f.Name())
	if err != nil {
		return nil, err
	}
	var p cveJiraPlan
	if err := json.Unmarshal(out, &p); err != nil {
		return nil, fmt.Errorf("parse plan: %w (raw: %s)", err, out)
	}
	return &p, nil
}

func discoverSprintID(ctx context.Context) int {
	out, err := runCveJira(ctx, "discover-sprint")
	if err != nil {
		return 0
	}
	var r struct {
		SprintID int `json:"sprint_id"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		return 0
	}
	return r.SprintID
}

// --- Handler ---------------------------------------------------------------

func (h *Handler) CreateCVEJiraTickets(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("JIRA_TOKEN") == "" {
		http.Error(w, "JIRA_TOKEN not configured", http.StatusServiceUnavailable)
		return
	}

	var req CreateCVEJiraTicketsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Detach Jira work from the request context so a Next.js proxy timeout
	// doesn't leave half the tickets created and half cancelled mid-flight.
	jiraCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	plan, err := planRepos(jiraCtx, req.Repos)
	if err != nil {
		http.Error(w, "plan failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sprintID := discoverSprintID(jiraCtx)
	monthLabel := time.Now().Format("2006-01")
	today := time.Now().Format("2006-01-02")

	resp := CreateCVEJiraTicketsResponse{
		Stories: map[string]JiraTicket{},
		Repos:   []JiraRepoTicket{},
		Errors:  map[string]string{},
	}

	storyKeys := map[string]string{}
	var mu sync.Mutex

	// Stories: one per component group, parallel.
	storyGroup, storyCtx := errgroup.WithContext(jiraCtx)
	storyGroup.SetLimit(8)
	for _, g := range plan.Groups {
		g := g
		storyGroup.Go(func() error {
			summary := fmt.Sprintf("CVE Remediation: %s — %s", g.Group, monthLabel)
			desc := fmt.Sprintf(
				"Parent story for CVE remediation of all %s repositories (%s).\n\nRepos:\n- %s",
				g.Group, today, strings.Join(g.Repos, "\n- "),
			)
			args := []string{
				"find-or-create-story",
				"--project", g.Project,
				"--summary", summary,
				"--description", desc,
				"--component", g.Component,
			}
			if sprintID > 0 {
				args = append(args, "--sprint-id", strconv.Itoa(sprintID))
			}
			t, err := callTicket(storyCtx, args...)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				slog.Warn("jira story create failed", "group", g.Group, "project", g.Project, "error", err.Error())
				resp.Errors["story:"+g.Group] = err.Error()
				return nil
			}
			storyKeys[g.Group] = t.Key
			resp.Stories[g.Group] = JiraTicket{Key: t.Key, URL: t.URL, Existing: t.Existing}
			return nil
		})
	}
	_ = storyGroup.Wait()

	// Repo → group lookup for the sub-task fan-out.
	repoMeta := map[string]cveJiraPlanGroup{}
	for _, g := range plan.Groups {
		for _, repo := range g.Repos {
			repoMeta[repo] = g
		}
	}

	// Sub-tasks: one per repo, parallel. Sprint inherited from parent — never set.
	subtaskGroup, subtaskCtx := errgroup.WithContext(jiraCtx)
	subtaskGroup.SetLimit(25)
	for _, repo := range req.Repos {
		repo := repo
		subtaskGroup.Go(func() error {
			meta := repoMeta[repo]
			mu.Lock()
			parentKey := storyKeys[meta.Group]
			mu.Unlock()

			summary := fmt.Sprintf("CVE Remediation: %s — %s", repo, monthLabel)
			desc := fmt.Sprintf(
				"CVE scan and remediation for `%s`.\n\nBranch: `chore/cve-remediation-%s`\nParent story: %s",
				repo, today, parentKey,
			)
			args := []string{
				"find-or-create-subtask",
				"--project", meta.Project,
				"--parent-key", parentKey,
				"--summary", summary,
				"--description", desc,
			}
			t, err := callTicket(subtaskCtx, args...)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				slog.Warn("jira subtask create failed", "repo", repo, "group", meta.Group, "project", meta.Project, "parent_key", parentKey, "error", err.Error())
				resp.Errors["subtask:"+repo] = err.Error()
				resp.Repos = append(resp.Repos, JiraRepoTicket{Repo: repo, Component: meta.Group, StoryKey: parentKey})
				return nil
			}
			resp.Repos = append(resp.Repos, JiraRepoTicket{
				Repo: repo, Component: meta.Group, StoryKey: parentKey,
				SubtaskKey: t.Key, SubtaskURL: t.URL, Existing: t.Existing,
			})
			return nil
		})
	}
	_ = subtaskGroup.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
