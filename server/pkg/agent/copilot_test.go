package agent

import (
	"log/slog"
	"strings"
	"testing"
)

func TestCopilotProcessEventsTextOnly(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)

	stream := strings.Join([]string{
		`{"type":"assistant.turn_start","data":{"turnId":"0","interactionId":"abc"}}`,
		`{"type":"assistant.message","data":{"messageId":"msg-1","content":"OK","toolRequests":[],"phase":"final_answer"}}`,
		`{"type":"result","sessionId":"session-123","exitCode":0}`,
	}, "\n")

	result := b.processEvents(strings.NewReader(stream), ch)

	if result.status != "completed" {
		t.Fatalf("status = %q, want completed", result.status)
	}
	if result.output != "OK" {
		t.Fatalf("output = %q, want OK", result.output)
	}
	if result.sessionID != "session-123" {
		t.Fatalf("sessionID = %q, want session-123", result.sessionID)
	}
	if len(ch) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(ch))
	}
	msg := <-ch
	if msg.Type != MessageStatus || msg.Status != "running" {
		t.Fatalf("first message = %#v, want running status", msg)
	}
	msg = <-ch
	if msg.Type != MessageText || msg.Content != "OK" {
		t.Fatalf("second message = %#v, want final text", msg)
	}
}

func TestCopilotProcessEventsToolExecutionAndDeltas(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)

	stream := strings.Join([]string{
		`{"type":"assistant.turn_start","data":{"turnId":"0","interactionId":"abc"}}`,
		`{"type":"assistant.message","data":{"messageId":"msg-tools","content":"","toolRequests":[{"toolCallId":"call-1","name":"bash","arguments":{"command":"pwd","description":"Print working directory"},"type":"function"}]}}`,
		`{"type":"tool.execution_start","data":{"toolCallId":"call-1","toolName":"bash","arguments":{"command":"pwd","description":"Print working directory"}}}`,
		`{"type":"tool.execution_complete","data":{"toolCallId":"call-1","success":true,"result":{"content":"/private/tmp\n<exited with exit code 0>","detailedContent":"/private/tmp\n<exited with exit code 0>"}}}`,
		`{"type":"assistant.turn_start","data":{"turnId":"1","interactionId":"abc"}}`,
		`{"type":"assistant.message_delta","data":{"messageId":"msg-final","deltaContent":"/private"}}`,
		`{"type":"assistant.message_delta","data":{"messageId":"msg-final","deltaContent":"/tmp"}}`,
		`{"type":"assistant.message","data":{"messageId":"msg-final","content":"/private/tmp","toolRequests":[],"phase":"final_answer"}}`,
		`{"type":"result","sessionId":"session-456","exitCode":0}`,
	}, "\n")

	result := b.processEvents(strings.NewReader(stream), ch)

	if result.status != "completed" {
		t.Fatalf("status = %q, want completed", result.status)
	}
	if result.output != "/private/tmp" {
		t.Fatalf("output = %q, want /private/tmp", result.output)
	}
	if result.sessionID != "session-456" {
		t.Fatalf("sessionID = %q, want session-456", result.sessionID)
	}
	if len(ch) != 6 {
		t.Fatalf("expected 6 messages, got %d", len(ch))
	}

	msg := <-ch
	if msg.Type != MessageStatus || msg.Status != "running" {
		t.Fatalf("message 1 = %#v, want running status", msg)
	}
	msg = <-ch
	if msg.Type != MessageToolUse || msg.Tool != "bash" || msg.CallID != "call-1" {
		t.Fatalf("message 2 = %#v, want tool use", msg)
	}
	msg = <-ch
	if msg.Type != MessageToolResult || !strings.Contains(msg.Output, "/private/tmp") {
		t.Fatalf("message 3 = %#v, want tool result", msg)
	}
	msg = <-ch
	if msg.Type != MessageStatus || msg.Status != "running" {
		t.Fatalf("message 4 = %#v, want second running status", msg)
	}
	msg = <-ch
	if msg.Type != MessageText || msg.Content != "/private" {
		t.Fatalf("message 5 = %#v, want first delta", msg)
	}
	msg = <-ch
	if msg.Type != MessageText || msg.Content != "/tmp" {
		t.Fatalf("message 6 = %#v, want second delta", msg)
	}
}

func TestCopilotProcessEventsError(t *testing.T) {
	t.Parallel()

	b := &copilotBackend{cfg: Config{Logger: slog.Default()}}
	ch := make(chan Message, 10)

	stream := strings.Join([]string{
		`{"type":"error","data":{"message":"Authentication required"}}`,
		`{"type":"result","sessionId":"session-789","exitCode":1}`,
	}, "\n")

	result := b.processEvents(strings.NewReader(stream), ch)

	if result.status != "failed" {
		t.Fatalf("status = %q, want failed", result.status)
	}
	if result.errMsg != "Authentication required" {
		t.Fatalf("errMsg = %q, want Authentication required", result.errMsg)
	}
	if len(ch) != 1 {
		t.Fatalf("expected 1 message, got %d", len(ch))
	}
	msg := <-ch
	if msg.Type != MessageError || msg.Content != "Authentication required" {
		t.Fatalf("error message = %#v", msg)
	}
}
