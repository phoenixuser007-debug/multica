package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	jiraBaseURL = "https://jira.arubanetworks.com"
	jiraProject = "CNX"
)

// componentGroup maps a repo name prefix to its human-readable group name (used as story title).
func componentGroup(repo string) string {
	switch {
	case strings.HasPrefix(repo, "bridge-5g-"):
		return "Bridge 5G"
	case strings.HasPrefix(repo, "edge-platform-"):
		return "Edge Platform"
	case strings.HasPrefix(repo, "edge-plugin-"):
		return "Edge Plugin"
	case strings.HasPrefix(repo, "iotops-client-location-"):
		return "Edge Location"
	case strings.HasPrefix(repo, "iotops-client-"):
		return "IoTOps Client"
	case strings.HasPrefix(repo, "iotops-"):
		return "IoTOps"
	case strings.HasPrefix(repo, "ui-"):
		return "UI"
	default:
		return "Other"
	}
}

// jiraComponent returns the valid CNX Jira component name for a given group.
func jiraComponent(group string) string {
	switch group {
	case "Bridge 5G":
		return "Bridge Monitoring"
	case "Edge Platform":
		return "Edge Platform"
	case "Edge Plugin":
		return "Edge App Config"
	case "Edge Location":
		return "Edge Location Engine"
	case "IoTOps Client":
		return "Unified Client - Non IP"
	case "IoTOps":
		return "IoT"
	case "UI":
		return "Edge Platform UI"
	default:
		return "Edge Platform"
	}
}

// JiraTicket holds a Jira issue key, URL, and whether it already existed.
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

func (h *Handler) CreateCVEJiraTickets(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("JIRA_TOKEN")
	if token == "" {
		http.Error(w, "JIRA_TOKEN not configured", http.StatusServiceUnavailable)
		return
	}

	var req CreateCVEJiraTicketsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	today := time.Now().Format("2006-01-02")
	monthLabel := currentJiraMonthLabel()
	fixVersion := currentJiraFixVersion()

	resp := CreateCVEJiraTicketsResponse{
		Stories: map[string]JiraTicket{},
		Repos:   []JiraRepoTicket{},
		Errors:  map[string]string{},
	}

	// Group repos by component
	groups := map[string][]string{}
	for _, repo := range req.Repos {
		cg := componentGroup(repo)
		groups[cg] = append(groups[cg], repo)
	}

	// Ensure one Story per component group for this month
	storyKeys := map[string]string{}
	for component, repos := range groups {
		storyTitle := fmt.Sprintf("CVE Remediation: %s", component)
		jql := fmt.Sprintf(
			`project = CNX AND issuetype = Story AND summary ~ "%s" AND summary ~ "%s" ORDER BY created DESC`,
			storyTitle, monthLabel,
		)
		existing, err := jiraSearchFirst(r.Context(), token, jql)
		if err == nil && existing != "" {
			storyKeys[component] = existing
			resp.Stories[component] = JiraTicket{Key: existing, URL: jiraWebURL(existing), Existing: true}
			continue
		}

		desc := fmt.Sprintf(
			"Parent story for CVE remediation of all %s repositories (%s).\n\nRepos:\n- %s",
			component, today, strings.Join(repos, "\n- "),
		)
		key, url, err := jiraCreate(r.Context(), token, jiraCreateFields{
			IssueType:   "Story",
			Summary:     fmt.Sprintf("%s — %s", storyTitle, monthLabel),
			Description: desc,
			Component:   jiraComponent(component),
			FixVersion:  fixVersion,
		})
		if err != nil {
			resp.Errors["story:"+component] = err.Error()
			continue
		}
		storyKeys[component] = key
		resp.Stories[component] = JiraTicket{Key: key, URL: url, Existing: false}
	}

	// Ensure one Sub-task per repo under its parent story
	for _, repo := range req.Repos {
		cg := componentGroup(repo)
		parentKey := storyKeys[cg]

		subtaskTitle := fmt.Sprintf("CVE Remediation: %s", repo)
		jql := fmt.Sprintf(
			`project = CNX AND issuetype = "Sub-task" AND summary ~ "%s" AND summary ~ "%s" ORDER BY created DESC`,
			subtaskTitle, monthLabel,
		)
		existing, err := jiraSearchFirst(r.Context(), token, jql)
		if err == nil && existing != "" {
			resp.Repos = append(resp.Repos, JiraRepoTicket{
				Repo: repo, Component: cg, StoryKey: parentKey,
				SubtaskKey: existing, SubtaskURL: jiraWebURL(existing), Existing: true,
			})
			continue
		}

		desc := fmt.Sprintf(
			"CVE scan and remediation for `%s`.\n\nBranch: `chore/cve-remediation-%s`\nParent story: %s",
			repo, today, parentKey,
		)
		key, url, err := jiraCreate(r.Context(), token, jiraCreateFields{
			IssueType:   "Sub-task",
			Summary:     fmt.Sprintf("%s — %s", subtaskTitle, monthLabel),
			Description: desc,
			ParentKey:   parentKey,
		})
		if err != nil {
			resp.Errors["subtask:"+repo] = err.Error()
			resp.Repos = append(resp.Repos, JiraRepoTicket{Repo: repo, Component: cg, StoryKey: parentKey})
			continue
		}
		resp.Repos = append(resp.Repos, JiraRepoTicket{
			Repo: repo, Component: cg, StoryKey: parentKey,
			SubtaskKey: key, SubtaskURL: url, Existing: false,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// jiraSearchFirst runs a JQL query against the Jira REST API and returns the first issue key.
func jiraSearchFirst(ctx context.Context, token, jql string) (string, error) {
	params := url.Values{}
	params.Set("jql", jql)
	params.Set("maxResults", "1")
	params.Set("fields", "summary")

	req, err := http.NewRequestWithContext(ctx, "GET",
		jiraBaseURL+"/rest/api/2/search?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("search %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		Total  int `json:"total"`
		Issues []struct {
			Key string `json:"key"`
		} `json:"issues"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Total == 0 || len(result.Issues) == 0 {
		return "", fmt.Errorf("not found")
	}
	return result.Issues[0].Key, nil
}

type jiraCreateFields struct {
	IssueType   string
	Summary     string
	Description string
	Component   string
	FixVersion  string
	ParentKey   string
}

// jiraCreate creates a Jira issue via the REST API and returns the new key + URL.
func jiraCreate(ctx context.Context, token string, f jiraCreateFields) (string, string, error) {
	fields := map[string]any{
		"project":     map[string]string{"key": jiraProject},
		"summary":     f.Summary,
		"description": f.Description,
		"issuetype":   map[string]string{"name": f.IssueType},
	}
	if f.Component != "" {
		fields["components"] = []map[string]string{{"name": f.Component}}
	}
	isSubtask := strings.EqualFold(f.IssueType, "sub-task")
	if f.FixVersion != "" && !isSubtask {
		fields["fixVersions"] = []map[string]string{{"name": f.FixVersion}}
		fields["versions"] = []map[string]string{{"name": f.FixVersion}}
	}
	if f.ParentKey != "" {
		fields["parent"] = map[string]string{"key": f.ParentKey}
	}
	if isSubtask {
		fields["customfield_12201"] = []map[string]string{{"value": "Cloud"}}
	}

	body, err := json.Marshal(map[string]any{"fields": fields})
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		jiraBaseURL+"/rest/api/2/issue", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("%d — %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil || result.Key == "" {
		return "", "", fmt.Errorf("unexpected response: %s", string(respBody))
	}
	return result.Key, jiraWebURL(result.Key), nil
}

func jiraWebURL(key string) string {
	return jiraBaseURL + "/browse/" + key
}

func currentJiraMonthLabel() string {
	return time.Now().Format("2006-01")
}

func currentJiraFixVersion() string {
	months := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	t := time.Now()
	return fmt.Sprintf("CNX-%s-%d", months[t.Month()-1], t.Year())
}
