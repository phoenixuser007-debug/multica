package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	wsRepoPath   = "/root/dev-env/ws"
	stashBaseURL = "ssh://git@stash.arubanetworks.com/gvt"
	stashBranch  = "devel"
)

// RepoStatusItem is the status of a single repository.
type RepoStatusItem struct {
	Name    string `json:"name"`
	Present bool   `json:"present"`
	Path    string `json:"path"`
}

// GetReposStatus checks which repos from the canonical CVE list are present on disk.
func (h *Handler) GetReposStatus(w http.ResponseWriter, r *http.Request) {
	items := make([]RepoStatusItem, len(cveRepoList))
	for i, name := range cveRepoList {
		path := filepath.Join(wsRepoPath, name)
		_, err := os.Stat(path)
		items[i] = RepoStatusItem{
			Name:    name,
			Present: err == nil,
			Path:    path,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// CloneReposRequest is the request body for cloning missing repos.
type CloneReposRequest struct {
	Repos []string `json:"repos"`
}

// CloneReposResponse reports which repos were cloned and which failed.
type CloneReposResponse struct {
	Cloned []string            `json:"cloned"`
	Failed map[string]string   `json:"failed"`
}

// CloneRepos clones repos that are not yet present in the workspace path.
func (h *Handler) CloneRepos(w http.ResponseWriter, r *http.Request) {
	var req CloneReposRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp := CloneReposResponse{
		Cloned: []string{},
		Failed: map[string]string{},
	}

	for _, name := range req.Repos {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		destPath := filepath.Join(wsRepoPath, name)
		// Skip if already present
		if _, err := os.Stat(destPath); err == nil {
			continue
		}

		cloneURL := stashBaseURL + "/" + name + ".git"
		cmd := exec.CommandContext(r.Context(), "git", "clone", "--depth=1", "-b", stashBranch, cloneURL, destPath)
		cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null")
		if out, err := cmd.CombinedOutput(); err != nil {
			resp.Failed[name] = strings.TrimSpace(string(out))
		} else {
			resp.Cloned = append(resp.Cloned, name)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// cveRepoList is the canonical list of repos subject to CVE remediation.
var cveRepoList = []string{
	"bridge-5g-ccs-interaction",
	"bridge-5g-cloud-proto",
	"bridge-5g-common-lib",
	"bridge-5g-device-connectivity",
	"bridge-5g-device-health",
	"bridge-5g-dia",
	"bridge-5g-events-processor",
	"bridge-5g-exporter-service",
	"bridge-5g-kafka",
	"bridge-5g-message-transformer",
	"bridge-5g-state-graphql-service",
	"bridge-5g-state-processor",
	"bridge-5g-state-publisher",
	"bridge-5g-stats-graphql-service",
	"bridge-5g-stats-processor",
	"bridge-5g-stats-publisher",
	"edge-platform-core",
	"edge-platform-debug-dashboard",
	"edge-platform-device-connectivity",
	"edge-platform-device-health",
	"edge-platform-events-processor",
	"edge-platform-graphql-service",
	"edge-platform-io",
	"edge-platform-message-transformer",
	"edge-platform-proto",
	"edge-platform-state-enricher",
	"edge-platform-state-processor",
	"edge-platform-state-publisher",
	"edge-plugin-graphql-service",
	"edge-plugin-message-transformer",
	"edge-plugin-sia",
	"edge-plugin-state-processor",
	"edge-plugin-state-publisher",
	"iotops-client-attributes-graphql-service",
	"iotops-client-attributes-processor",
	"iotops-client-attributes-publisher",
	"iotops-client-location-aggregator",
	"iotops-client-location-ap-syncher",
	"iotops-client-location-engine",
	"iotops-client-location-message-transformer",
	"iotops-client-location-processor",
	"iotops-client-location-rssi-transformer",
	"iotops-client-message-transformer",
	"iotops-client-state-graphql-service",
	"iotops-client-state-processor",
	"iotops-client-state-publisher",
	"iotops-client-stats-graphql-service",
	"iotops-client-stats-processor",
	"iotops-client-stats-publisher",
	"iotops-cloud-proto",
	"iotops-graphql-service",
	"iotops-message-transformer",
	"iotops-state-processor",
	"iotops-state-publisher",
	"iotops-stats-processor",
	"iotops-stats-publisher",
	"ui-edge-platform-management",
}
