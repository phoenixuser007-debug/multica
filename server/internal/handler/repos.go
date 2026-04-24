package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/multica-ai/multica/server/internal/middleware"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

const (
	wsRepoPath        = "/root/dev-env/ws"
	centralRepoPath   = "/root/central"
	stashGvtBaseURL   = "ssh://git@stash.arubanetworks.com/gvt"
	stashIIBaseURL    = "ssh://git@stash.arubanetworks.com/ii"
	stashBranchGvt    = "devel"
	stashBranchII     = "release/ii_1.118.0"
)

// repoFamily resolves a repo name to its source-of-truth location and branch.
// Central (ii-ae-*) repos live in /root/central and clone from the `ii` Stash project
// on the release/ii_1.118.0 branch. Everything else lives in /root/dev-env/ws and
// clones from the `gvt` Stash project on devel.
type repoFamily struct {
	HostBase string
	StashURL string
	Branch   string
}

func familyFor(repo string) repoFamily {
	if strings.HasPrefix(repo, "ii-ae-") {
		return repoFamily{HostBase: centralRepoPath, StashURL: stashIIBaseURL, Branch: stashBranchII}
	}
	return repoFamily{HostBase: wsRepoPath, StashURL: stashGvtBaseURL, Branch: stashBranchGvt}
}

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
		fam := familyFor(name)
		path := filepath.Join(fam.HostBase, name)
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

// CloneRepos clones repos that are not yet present on disk.
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
		fam := familyFor(name)
		destPath := filepath.Join(fam.HostBase, name)
		// Skip if already present
		if _, err := os.Stat(destPath); err == nil {
			continue
		}

		cloneURL := fam.StashURL + "/" + name + ".git"
		cmd := exec.CommandContext(r.Context(), "git", "clone", "--depth=1", "-b", fam.Branch, cloneURL, destPath)
		cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null")
		if out, err := cmd.CombinedOutput(); err != nil {
			resp.Failed[name] = strings.TrimSpace(string(out))
		} else {
			resp.Cloned = append(resp.Cloned, name)
		}
	}

	// Ensure every requested repo is registered in workspace.repos so the
	// agent's daemon discovers it. Without this the agent reports "no repos
	// in workspace" and never starts work even though the clone succeeded.
	if wsID := middleware.WorkspaceIDFromContext(r.Context()); wsID != "" {
		if err := h.ensureWorkspaceRepos(r.Context(), wsID, req.Repos); err != nil {
			slog.Warn("ensure workspace repos failed", "error", err, "workspace_id", wsID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ensureWorkspaceRepos appends any of the named repos that aren't already in
// the workspace's repo list, using each repo's family to derive the Stash URL.
// Existing entries are kept untouched. No-op when nothing needs adding.
func (h *Handler) ensureWorkspaceRepos(ctx context.Context, workspaceID string, repos []string) error {
	ws, err := h.Queries.GetWorkspace(ctx, parseUUID(workspaceID))
	if err != nil {
		return err
	}
	current := parseWorkspaceRepos(ws.Repos)
	known := make(map[string]bool, len(current))
	for _, r := range current {
		known[r.URL] = true
	}

	added := false
	for _, name := range repos {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		fam := familyFor(name)
		url := fam.StashURL + "/" + name + ".git"
		if known[url] {
			continue
		}
		current = append(current, RepoData{URL: url, Description: name})
		known[url] = true
		added = true
	}
	if !added {
		return nil
	}

	reposJSON, err := json.Marshal(current)
	if err != nil {
		return err
	}
	_, err = h.Queries.UpdateWorkspace(ctx, db.UpdateWorkspaceParams{
		ID:    parseUUID(workspaceID),
		Repos: reposJSON,
	})
	return err
}

// cveRepoList is the canonical list of repos subject to CVE remediation.
// gvt (devel branch, /root/dev-env/ws) — bridge-5g, edge-platform, edge-plugin, iotops, ui
// ii  (release/ii_1.118.0, /root/central) — ii-ae-* (Maven-only; non-Maven ii-ae-* are skipped)
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
	"ii-ae-api-gateway",
	"ii-ae-appstore",
	"ii-ae-configuration",
	"ii-ae-devices",
	"ii-ae-iot-gateway",
	"ii-ae-location",
	"ii-ae-location-engine-feeder",
	"ii-ae-message-transformer",
	"ii-ae-radio",
	"ii-ae-transport-profile",
}
