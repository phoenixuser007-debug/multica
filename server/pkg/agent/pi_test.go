package agent

import (
	"log/slog"
	"reflect"
	"testing"
)

func TestNewReturnsPiBackend(t *testing.T) {
	t.Parallel()
	b, err := New("pi", Config{ExecutablePath: "/nonexistent/pi"})
	if err != nil {
		t.Fatalf("New(pi) error: %v", err)
	}
	if _, ok := b.(*piBackend); !ok {
		t.Fatalf("expected *piBackend, got %T", b)
	}
}

func TestBuildPiArgsMatchesBadlogicCodingAgentJSONMode(t *testing.T) {
	t.Parallel()

	args := buildPiArgs("fix it", "/tmp/session.jsonl", ExecOptions{
		Model:        "openai/gpt-4o",
		SystemPrompt: "extra rules",
		CustomArgs:   []string{"--thinking", "high", "--mode", "interactive"},
	}, slog.Default())

	want := []string{
		"-p",
		"--mode", "json",
		"--session", "/tmp/session.jsonl",
		"--provider", "openai",
		"--model", "gpt-4o",
		"--tools", "read,bash,edit,write,grep,find,ls",
		"--append-system-prompt", "extra rules",
		"--thinking", "high",
		"fix it",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("buildPiArgs() = %#v, want %#v", args, want)
	}
}
