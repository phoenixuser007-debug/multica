package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// copilotBackend implements Backend by spawning `copilot -p ... --output-format json`
// and reading JSONL events from stdout.
type copilotBackend struct {
	cfg Config
}

func (b *copilotBackend) Execute(ctx context.Context, prompt string, opts ExecOptions) (*Session, error) {
	execPath := b.cfg.ExecutablePath
	if execPath == "" {
		execPath = "copilot"
	}
	if _, err := exec.LookPath(execPath); err != nil {
		return nil, fmt.Errorf("copilot executable not found at %q: %w", execPath, err)
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 20 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)

	args := []string{
		"-p", prompt,
		"--output-format", "json",
		"--stream", "on",
		"--allow-all-tools",
		"--allow-all-paths",
		"--allow-all-urls",
		"--no-ask-user",
		"--no-auto-update",
	}
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	if opts.ResumeSessionID != "" {
		args = append(args, "--resume="+opts.ResumeSessionID)
	}
	if opts.SystemPrompt != "" {
		b.cfg.Logger.Warn("copilot does not expose a direct system-prompt flag; relying on repository instructions instead")
	}

	cmd := exec.CommandContext(runCtx, execPath, args...)
	if opts.Cwd != "" {
		cmd.Dir = opts.Cwd
	}
	cmd.Env = buildEnv(b.cfg.Env)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("copilot stdout pipe: %w", err)
	}
	cmd.Stderr = newLogWriter(b.cfg.Logger, "[copilot:stderr] ")

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start copilot: %w", err)
	}

	b.cfg.Logger.Info("copilot started", "pid", cmd.Process.Pid, "cwd", opts.Cwd, "model", opts.Model)

	msgCh := make(chan Message, 256)
	resCh := make(chan Result, 1)

	go func() {
		defer cancel()
		defer close(msgCh)
		defer close(resCh)

		startTime := time.Now()
		scanResult := b.processEvents(stdout, msgCh)

		exitErr := cmd.Wait()
		duration := time.Since(startTime)

		if runCtx.Err() == context.DeadlineExceeded {
			scanResult.status = "timeout"
			scanResult.errMsg = fmt.Sprintf("copilot timed out after %s", timeout)
		} else if runCtx.Err() == context.Canceled {
			scanResult.status = "aborted"
			scanResult.errMsg = "execution cancelled"
		} else if exitErr != nil && scanResult.status == "completed" {
			scanResult.status = "failed"
			scanResult.errMsg = fmt.Sprintf("copilot exited with error: %v", exitErr)
		}

		b.cfg.Logger.Info("copilot finished", "pid", cmd.Process.Pid, "status", scanResult.status, "duration", duration.Round(time.Millisecond).String())

		resCh <- Result{
			Status:     scanResult.status,
			Output:     scanResult.output,
			Error:      scanResult.errMsg,
			DurationMs: duration.Milliseconds(),
			SessionID:  scanResult.sessionID,
		}
	}()

	return &Session{Messages: msgCh, Result: resCh}, nil
}

type copilotEventResult struct {
	status    string
	errMsg    string
	output    string
	sessionID string
}

func (b *copilotBackend) processEvents(r io.Reader, ch chan<- Message) copilotEventResult {
	finalStatus := "completed"
	var finalError string
	var output strings.Builder
	var sessionID string
	streamedMessages := make(map[string]bool)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event copilotEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		if event.SessionID != "" {
			sessionID = event.SessionID
		}

		switch event.Type {
		case "assistant.turn_start":
			trySend(ch, Message{Type: MessageStatus, Status: "running"})
		case "assistant.message_delta":
			var data copilotMessageDeltaData
			if err := json.Unmarshal(event.Data, &data); err != nil {
				continue
			}
			if data.MessageID != "" {
				streamedMessages[data.MessageID] = true
			}
			if data.DeltaContent != "" {
				output.WriteString(data.DeltaContent)
				trySend(ch, Message{Type: MessageText, Content: data.DeltaContent})
			}
		case "assistant.message":
			var data copilotAssistantMessageData
			if err := json.Unmarshal(event.Data, &data); err != nil {
				continue
			}
			if data.Content != "" && !streamedMessages[data.MessageID] {
				output.WriteString(data.Content)
				trySend(ch, Message{Type: MessageText, Content: data.Content})
			}
		case "assistant.reasoning":
			var data copilotReasoningData
			if err := json.Unmarshal(event.Data, &data); err != nil {
				continue
			}
			if data.Content != "" {
				trySend(ch, Message{Type: MessageThinking, Content: data.Content})
			}
		case "tool.execution_start":
			var data copilotToolExecutionStartData
			if err := json.Unmarshal(event.Data, &data); err != nil {
				continue
			}
			trySend(ch, Message{
				Type:   MessageToolUse,
				Tool:   data.ToolName,
				CallID: data.ToolCallID,
				Input:  data.Arguments,
			})
		case "tool.execution_complete":
			var data copilotToolExecutionCompleteData
			if err := json.Unmarshal(event.Data, &data); err != nil {
				continue
			}
			outputStr := data.Result.DetailedContent
			if outputStr == "" {
				outputStr = data.Result.Content
			}
			trySend(ch, Message{
				Type:   MessageToolResult,
				CallID: data.ToolCallID,
				Output: outputStr,
			})
		case "error", "session.error", "assistant.error":
			errMsg := extractCopilotError(event.Data)
			if errMsg == "" {
				errMsg = "unknown copilot error"
			}
			finalStatus = "failed"
			finalError = errMsg
			trySend(ch, Message{Type: MessageError, Content: errMsg})
		case "result":
			if event.ExitCode != 0 && finalStatus == "completed" {
				finalStatus = "failed"
				if finalError == "" {
					finalError = fmt.Sprintf("copilot exited with code %d", event.ExitCode)
				}
			}
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		b.cfg.Logger.Warn("copilot stdout scanner error", "error", scanErr)
		if finalStatus == "completed" {
			finalStatus = "failed"
			finalError = fmt.Sprintf("stdout read error: %v", scanErr)
		}
	}

	return copilotEventResult{
		status:    finalStatus,
		errMsg:    finalError,
		output:    output.String(),
		sessionID: sessionID,
	}
}

func extractCopilotError(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var payload struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(raw, &payload); err == nil {
		if payload.Message != "" {
			return payload.Message
		}
		if payload.Error != "" {
			return payload.Error
		}
	}

	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return ""
	}
	for _, key := range []string{"message", "error"} {
		if s, ok := generic[key].(string); ok && s != "" {
			return s
		}
	}
	return ""
}

type copilotEvent struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	SessionID string          `json:"sessionId"`
	ExitCode  int             `json:"exitCode"`
}

type copilotMessageDeltaData struct {
	MessageID    string `json:"messageId"`
	DeltaContent string `json:"deltaContent"`
}

type copilotAssistantMessageData struct {
	MessageID string `json:"messageId"`
	Content   string `json:"content"`
}

type copilotReasoningData struct {
	Content string `json:"content"`
}

type copilotToolExecutionStartData struct {
	ToolCallID string         `json:"toolCallId"`
	ToolName   string         `json:"toolName"`
	Arguments  map[string]any `json:"arguments"`
}

type copilotToolExecutionCompleteData struct {
	ToolCallID string `json:"toolCallId"`
	Result     struct {
		Content         string `json:"content"`
		DetailedContent string `json:"detailedContent"`
	} `json:"result"`
}
