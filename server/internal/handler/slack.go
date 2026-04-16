package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// SlackNotifyRequest is the payload sent by the agent when a CVE ticket is done.
type SlackNotifyRequest struct {
	Repo             string `json:"repo"`
	JiraKey          string `json:"jira_key"`
	JiraURL          string `json:"jira_url"`
	PRTitle          string `json:"pr_title"`
	PRURL            string `json:"pr_url"`
	CVEHighBefore    int    `json:"cve_high_before"`
	CVECriticalBefore int   `json:"cve_critical_before"`
	CVEHighAfter     int    `json:"cve_high_after"`
	CVECriticalAfter int    `json:"cve_critical_after"`
}

// NotifyCVEDone posts a Slack message to #prs-trivy when a CVE remediation
// ticket is moved to done and a PR has been raised.
func (h *Handler) NotifyCVEDone(w http.ResponseWriter, r *http.Request) {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		http.Error(w, "SLACK_WEBHOOK_URL not configured", http.StatusServiceUnavailable)
		return
	}

	var req SlackNotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	msg := buildSlackMessage(req)
	if err := postSlackWebhook(webhookURL, msg); err != nil {
		http.Error(w, fmt.Sprintf("slack post failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func buildSlackMessage(req SlackNotifyRequest) map[string]any {
	var lines []string

	lines = append(lines, fmt.Sprintf("*CVE Remediation done:* `%s`", req.Repo))

	if req.JiraKey != "" && req.JiraURL != "" {
		lines = append(lines, fmt.Sprintf("*Jira:* <%s|%s>", req.JiraURL, req.JiraKey))
	}
	if req.PRURL != "" {
		title := req.PRTitle
		if title == "" {
			title = "View PR"
		}
		lines = append(lines, fmt.Sprintf("*PR:* <%s|%s>", req.PRURL, title))
	}

	// CVE summary
	highRemoved := req.CVEHighBefore - req.CVEHighAfter
	critRemoved := req.CVECriticalBefore - req.CVECriticalAfter
	if req.CVEHighBefore > 0 || req.CVECriticalBefore > 0 {
		lines = append(lines, fmt.Sprintf(
			"*CVEs removed:* :red_circle: Critical %d→%d (-%d)  :large_orange_circle: High %d→%d (-%d)",
			req.CVECriticalBefore, req.CVECriticalAfter, critRemoved,
			req.CVEHighBefore, req.CVEHighAfter, highRemoved,
		))
	} else {
		lines = append(lines, "*CVEs:* No HIGH or CRITICAL found")
	}

	return map[string]any{
		"text": ":white_check_mark: CVE remediation complete — " + req.Repo,
		"blocks": []map[string]any{
			{
				"type": "section",
				"text": map[string]any{
					"type": "mrkdwn",
					"text": strings.Join(lines, "\n"),
				},
			},
		},
	}
}

func postSlackWebhook(webhookURL string, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned %d", resp.StatusCode)
	}
	return nil
}
