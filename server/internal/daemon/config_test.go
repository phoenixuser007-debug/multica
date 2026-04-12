package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDetectsCopilotBinary(t *testing.T) {
	tmp := t.TempDir()
	if err := writeExecutable(filepath.Join(tmp, "copilot"), "#!/bin/sh\nexit 0\n"); err != nil {
		t.Fatalf("write fake copilot: %v", err)
	}

	t.Setenv("PATH", tmp)
	t.Setenv("MULTICA_WORKSPACES_ROOT", tmp)
	t.Setenv("MULTICA_COPILOT_MODEL", "gpt-5.3-codex")

	cfg, err := LoadConfig(Overrides{})
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	entry, ok := cfg.Agents["copilot"]
	if !ok {
		t.Fatalf("expected copilot agent entry, got %#v", cfg.Agents)
	}
	if entry.Path != "copilot" {
		t.Fatalf("path = %q, want copilot", entry.Path)
	}
	if entry.Model != "gpt-5.3-codex" {
		t.Fatalf("model = %q, want gpt-5.3-codex", entry.Model)
	}
}

func writeExecutable(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o755)
}
