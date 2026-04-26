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

func TestIsPiSessionPath(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"ses_242a1fe82ffe7TFk0FFyFVXrD1", false},        // Copilot-style opaque id
		{"abc123-def4-5678-9abc-def012345678", false},   // Claude-style UUID
		{"./local.jsonl", false},                          // not absolute
		{"/root/.multica/pi-sessions/abc.txt", false},     // wrong suffix
		{"/root/.multica/pi-sessions/20260426T191500.000000000.jsonl", true},
	}
	for _, c := range cases {
		if got := isPiSessionPath(c.in); got != c.want {
			t.Errorf("isPiSessionPath(%q) = %v, want %v", c.in, got, c.want)
		}
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
