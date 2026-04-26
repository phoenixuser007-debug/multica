package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// ACP JSON-RPC 2.0 transport shared by providers that speak the protocol.

type acpPromptResult struct {
	stopReason string
	usage      TokenUsage
}

type acpClient struct {
	cfg          Config
	stdin        interface{ Write([]byte) (int, error) }
	writeMu      sync.Mutex // serialises stdin.Write calls across goroutines
	mu           sync.Mutex
	nextID       int
	pending      map[int]*pendingRPC
	sessionID    string
	onMessage    func(Message)
	onPromptDone func(acpPromptResult)

	// pendingTools buffers the args for tool calls whose input streams in
	// across multiple ACP tool_call_update messages (kimi does this —
	// tokens from the LLM arrive one at a time, and each update carries
	// the cumulative args JSON so far). We defer emitting MessageToolUse
	// until we either see status=completed/failed or have a full arg set,
	// so the UI never sees a half-written command like `{"comma`.
	toolMu       sync.Mutex
	pendingTools map[string]*pendingToolCall

	usageMu sync.Mutex
	usage   TokenUsage
}

// pendingToolCall buffers state for a tool call while its arguments
// are streaming in. One entry per ACP toolCallId.
type pendingToolCall struct {
	toolName string         // already mapped via acpToolNameFromTitle
	input    map[string]any // from rawInput when the agent sends it up front (Kimi)
	argsText string         // accumulated `content[].text` args (kimi, cumulative)
	emitted  bool           // whether we've already sent MessageToolUse
}

// writeLine serialises concurrent JSON-RPC writes so request() (main
// goroutine) and handleAgentRequest() (reader goroutine) don't
// interleave frames. The pipe itself is atomic for small writes, but
// we also want deterministic ordering under contention.
func (c *acpClient) writeLine(data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_, err := c.stdin.Write(data)
	return err
}

func (c *acpClient) request(ctx context.Context, method string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	pr := &pendingRPC{ch: make(chan rpcResult, 1), method: method}
	c.pending[id] = pr
	c.mu.Unlock()

	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, err
	}
	data = append(data, '\n')
	if err := c.writeLine(data); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("write %s: %w", method, err)
	}

	select {
	case res := <-pr.ch:
		return res.result, res.err
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	}
}

func (c *acpClient) closeAllPending(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, pr := range c.pending {
		pr.ch <- rpcResult{err: err}
		delete(c.pending, id)
	}
}

func (c *acpClient) handleLine(line string) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return
	}

	// Agent → client request: has id + method (no result / error yet).
	// Kimi uses this for session/request_permission; if we don't answer,
	// the agent blocks for 300s and the task hangs.
	if _, hasID := raw["id"]; hasID {
		if _, hasResult := raw["result"]; hasResult {
			c.handleResponse(raw)
			return
		}
		if _, hasError := raw["error"]; hasError {
			c.handleResponse(raw)
			return
		}
		if _, hasMethod := raw["method"]; hasMethod {
			c.handleAgentRequest(raw)
			return
		}
	}

	// Notification (no id, has method) — session updates from the agent.
	if _, hasMethod := raw["method"]; hasMethod {
		c.handleNotification(raw)
	}
}

// handleAgentRequest replies to JSON-RPC requests the agent sends
// us (agent → client direction). The only one we care about today is
// `session/request_permission`: the daemon is headless and cannot
// actually prompt a user, so we auto-approve every action. Using
// `approve_for_session` rather than `approve` means subsequent
// identical actions (every Shell invocation, every file write) don't
// round-trip through us — the agent remembers them locally.
func (c *acpClient) handleAgentRequest(raw map[string]json.RawMessage) {
	var method string
	_ = json.Unmarshal(raw["method"], &method)

	rawID, ok := raw["id"]
	if !ok {
		return
	}

	var resp map[string]any
	switch method {
	case "session/request_permission":
		resp = map[string]any{
			"jsonrpc": "2.0",
			"id":      json.RawMessage(rawID),
			"result": map[string]any{
				"outcome": map[string]any{
					"outcome":  "selected",
					"optionId": "approve_for_session",
				},
			},
		}
		c.cfg.Logger.Debug("auto-approved agent permission request", "method", method)
	default:
		// Unknown agent→client method — reply with standard "method
		// not found" so the agent doesn't block waiting for us. Better
		// than silence: the agent can decide how to proceed.
		resp = map[string]any{
			"jsonrpc": "2.0",
			"id":      json.RawMessage(rawID),
			"error": map[string]any{
				"code":    -32601,
				"message": "method not found: " + method,
			},
		}
		c.cfg.Logger.Debug("unhandled agent→client request", "method", method)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		c.cfg.Logger.Warn("marshal agent-request response", "method", method, "error", err)
		return
	}
	data = append(data, '\n')
	if err := c.writeLine(data); err != nil {
		c.cfg.Logger.Warn("write agent-request response", "method", method, "error", err)
	}
}

func (c *acpClient) handleResponse(raw map[string]json.RawMessage) {
	var id int
	if err := json.Unmarshal(raw["id"], &id); err != nil {
		// Try float (JSON numbers are floats by default).
		var fid float64
		if err := json.Unmarshal(raw["id"], &fid); err != nil {
			return
		}
		id = int(fid)
	}

	c.mu.Lock()
	pr, ok := c.pending[id]
	if ok {
		delete(c.pending, id)
	}
	c.mu.Unlock()

	if !ok {
		return
	}

	if errData, hasErr := raw["error"]; hasErr {
		var rpcErr struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(errData, &rpcErr)
		pr.ch <- rpcResult{err: fmt.Errorf("%s: %s (code=%d)", pr.method, rpcErr.Message, rpcErr.Code)}
	} else {
		// If this is a prompt response, extract usage and stop reason.
		if pr.method == "session/prompt" {
			c.extractPromptResult(raw["result"])
		}
		pr.ch <- rpcResult{result: raw["result"]}
	}
}

func (c *acpClient) extractPromptResult(data json.RawMessage) {
	var resp struct {
		StopReason string `json:"stopReason"`
		Usage      *struct {
			InputTokens      int64 `json:"inputTokens"`
			OutputTokens     int64 `json:"outputTokens"`
			TotalTokens      int64 `json:"totalTokens"`
			ThoughtTokens    int64 `json:"thoughtTokens"`
			CachedReadTokens int64 `json:"cachedReadTokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}

	pr := acpPromptResult{
		stopReason: resp.StopReason,
	}
	if resp.Usage != nil {
		pr.usage = TokenUsage{
			InputTokens:     resp.Usage.InputTokens,
			OutputTokens:    resp.Usage.OutputTokens,
			CacheReadTokens: resp.Usage.CachedReadTokens,
		}
	}

	if c.onPromptDone != nil {
		c.onPromptDone(pr)
	}
}

func (c *acpClient) handleNotification(raw map[string]json.RawMessage) {
	var method string
	_ = json.Unmarshal(raw["method"], &method)

	if method != "session/update" {
		return
	}

	var params struct {
		SessionID string          `json:"sessionId"`
		Update    json.RawMessage `json:"update"`
	}
	if p, ok := raw["params"]; ok {
		_ = json.Unmarshal(p, &params)
	}
	if len(params.Update) == 0 {
		return
	}

	// Parse the update discriminator.
	var updateType struct {
		SessionUpdate string `json:"sessionUpdate"`
	}
	_ = json.Unmarshal(params.Update, &updateType)

	switch updateType.SessionUpdate {
	case "agent_message_chunk":
		c.handleAgentMessage(params.Update)
	case "agent_thought_chunk":
		c.handleAgentThought(params.Update)
	case "tool_call":
		c.handleToolCallStart(params.Update)
	case "tool_call_update":
		c.handleToolCallUpdate(params.Update)
	case "usage_update":
		c.handleUsageUpdate(params.Update)
	}
}

func (c *acpClient) handleAgentMessage(data json.RawMessage) {
	var msg struct {
		Content struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(data, &msg); err != nil || msg.Content.Text == "" {
		return
	}
	if c.onMessage != nil {
		c.onMessage(Message{Type: MessageText, Content: msg.Content.Text})
	}
}

func (c *acpClient) handleAgentThought(data json.RawMessage) {
	var msg struct {
		Content struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(data, &msg); err != nil || msg.Content.Text == "" {
		return
	}
	if c.onMessage != nil {
		c.onMessage(Message{Type: MessageThinking, Content: msg.Content.Text})
	}
}

func (c *acpClient) handleToolCallStart(data json.RawMessage) {
	var msg struct {
		ToolCallID string            `json:"toolCallId"`
		Title      string            `json:"title"`
		Kind       string            `json:"kind"`
		RawInput   map[string]any    `json:"rawInput"`
		Content    []json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	toolName := acpToolNameFromTitle(msg.Title, msg.Kind)

	// Kimi pre-populates rawInput on the initial tool_call — emit
	// MessageToolUse immediately so the UI can show the tool invocation
	// live. Record the emission so handleToolCallUpdate doesn't re-emit
	// on completion.
	if msg.RawInput != nil {
		c.trackTool(msg.ToolCallID, &pendingToolCall{
			toolName: toolName,
			input:    msg.RawInput,
			emitted:  true,
		})
		if c.onMessage != nil {
			c.onMessage(Message{
				Type:   MessageToolUse,
				Tool:   toolName,
				CallID: msg.ToolCallID,
				Input:  msg.RawInput,
			})
		}
		return
	}

	// Kimi streams args token-by-token across tool_call_update messages;
	// the initial tool_call often carries an empty content block. Buffer
	// the tool and defer MessageToolUse emission to avoid the UI seeing
	// a command with `{""` as its input.
	c.trackTool(msg.ToolCallID, &pendingToolCall{
		toolName: toolName,
		argsText: extractACPToolCallText(msg.Content),
		emitted:  false,
	})
}

func (c *acpClient) handleToolCallUpdate(data json.RawMessage) {
	var msg struct {
		ToolCallID string            `json:"toolCallId"`
		Status     string            `json:"status"`
		Title      string            `json:"title"`
		Kind       string            `json:"kind"`
		RawInput   map[string]any    `json:"rawInput"`
		RawOutput  string            `json:"rawOutput"`
		Content    []json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	// Mid-stream: only buffer updates. Kimi emits many of these per
	// tool call, each carrying the cumulative args JSON so far.
	if msg.Status != "completed" && msg.Status != "failed" {
		if pending := c.getPendingTool(msg.ToolCallID); pending != nil && !pending.emitted {
			if text := extractACPToolCallText(msg.Content); text != "" {
				// kimi streams the full cumulative args on every frame;
				// overwrite rather than concatenate.
				pending.argsText = text
			}
		}
		return
	}

	// Completion: emit any deferred MessageToolUse first, then the result.
	pending := c.takePendingTool(msg.ToolCallID)
	c.emitDeferredToolUse(pending, msg.ToolCallID, msg.Title, msg.Kind, msg.RawInput)

	output := msg.RawOutput
	if output == "" {
		output = extractACPToolCallText(msg.Content)
	}
	if c.onMessage != nil {
		c.onMessage(Message{
			Type:   MessageToolResult,
			CallID: msg.ToolCallID,
			Output: output,
		})
	}
}

// trackTool stores pending-tool state for a given callID. Lazy-inits
// the map so zero-value acpClient values (common in tests) don't
// panic on the first tool call.
func (c *acpClient) trackTool(callID string, p *pendingToolCall) {
	c.toolMu.Lock()
	defer c.toolMu.Unlock()
	if c.pendingTools == nil {
		c.pendingTools = make(map[string]*pendingToolCall)
	}
	c.pendingTools[callID] = p
}

// getPendingTool returns the pending entry (may be nil) without
// removing it. Safe to call on a zero-value acpClient.
func (c *acpClient) getPendingTool(callID string) *pendingToolCall {
	c.toolMu.Lock()
	defer c.toolMu.Unlock()
	if c.pendingTools == nil {
		return nil
	}
	return c.pendingTools[callID]
}

// takePendingTool removes and returns the pending entry, or nil if
// none was tracked (e.g. the tool completed before we saw its start,
// or we missed the start frame).
func (c *acpClient) takePendingTool(callID string) *pendingToolCall {
	c.toolMu.Lock()
	defer c.toolMu.Unlock()
	if c.pendingTools == nil {
		return nil
	}
	p := c.pendingTools[callID]
	delete(c.pendingTools, callID)
	return p
}

// emitDeferredToolUse emits a buffered MessageToolUse right before the
// matching MessageToolResult. Handles three cases:
//   - Kimi tool: already emitted on tool_call → skip
//   - kimi tool with streamed args → parse accumulated JSON as Input
//   - unknown tool (completed arrived without a start frame) →
//     synthesize minimal info from the update's own fields
func (c *acpClient) emitDeferredToolUse(
	p *pendingToolCall,
	callID, updateTitle, updateKind string,
	updateRawInput map[string]any,
) {
	if p != nil && p.emitted {
		return
	}

	var toolName string
	var input map[string]any

	switch {
	case p != nil && p.input != nil:
		// Pre-buffered rawInput path — shouldn't happen because we set
		// emitted=true in that case, but handle defensively.
		toolName = p.toolName
		input = p.input
	case p != nil:
		toolName = p.toolName
		input = parseToolArgsJSON(p.argsText)
	default:
		// No record of the start frame — fall back to the update's own
		// title/kind/rawInput so the UI at least sees the tool name.
		toolName = acpToolNameFromTitle(updateTitle, updateKind)
		input = updateRawInput
	}

	if c.onMessage == nil {
		return
	}
	c.onMessage(Message{
		Type:   MessageToolUse,
		Tool:   toolName,
		CallID: callID,
		Input:  input,
	})
}

// parseToolArgsJSON turns kimi's accumulated args string into the
// structured map the UI expects under Message.Input. Kimi sends args
// as a JSON-encoded object (`{"command":"echo hi"}`), so a full JSON
// parse recovers the original tool-arg shape. On malformed input
// (streaming glitch, non-JSON tool) we preserve the raw text under a
// `text` key so the UI still has something to render.
func parseToolArgsJSON(argsText string) map[string]any {
	argsText = strings.TrimSpace(argsText)
	if argsText == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(argsText), &m); err == nil {
		return m
	}
	return map[string]any{"text": argsText}
}

// extractACPToolCallText concatenates the rendered text of every ACP
// block in a tool_call / tool_call_update's `content` array.
//
// Handles the two block types kimi emits:
//   - {type:"content", content:{type:"text", text:"..."}} — plain text
//     (shell output, tool args). Text is concatenated verbatim.
//   - {type:"diff", path, oldText, newText} — FileEdit output. Rendered
//     as a minimal unified-diff header so the UI distinguishes writes
//     from reads without needing a diff viewer.
//
// Terminal blocks ({type:"terminal", terminalId}) reference a remote
// terminal the client would normally subscribe to via terminal/output;
// we don't advertise terminal capability so we never receive those in
// practice, but if one slips through we skip it (nothing useful to
// surface from a bare ID).
func extractACPToolCallText(blocks []json.RawMessage) string {
	var b strings.Builder
	appendPiece := func(piece string) {
		if piece == "" {
			return
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(piece)
	}
	for _, raw := range blocks {
		var kind struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &kind); err != nil {
			continue
		}
		switch kind.Type {
		case "content":
			var outer struct {
				Content json.RawMessage `json:"content"`
			}
			if err := json.Unmarshal(raw, &outer); err != nil || len(outer.Content) == 0 {
				continue
			}
			var inner struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}
			if err := json.Unmarshal(outer.Content, &inner); err != nil {
				continue
			}
			if inner.Type != "text" {
				continue
			}
			appendPiece(inner.Text)
		case "diff":
			var diff struct {
				Path    string `json:"path"`
				OldText string `json:"oldText"`
				NewText string `json:"newText"`
			}
			if err := json.Unmarshal(raw, &diff); err != nil || diff.Path == "" {
				continue
			}
			// Keep it tiny — a full unified diff can be huge and we're
			// really just recording "this tool wrote to this file".
			// The UI can re-read the file if it needs the actual content.
			var piece strings.Builder
			piece.WriteString("--- ")
			piece.WriteString(diff.Path)
			piece.WriteString("\n+++ ")
			piece.WriteString(diff.Path)
			if diff.OldText == "" {
				piece.WriteString("\n(new file, ")
				piece.WriteString(strconv.Itoa(len(diff.NewText)))
				piece.WriteString(" bytes)")
			} else {
				piece.WriteString("\n(edited: ")
				piece.WriteString(strconv.Itoa(len(diff.OldText)))
				piece.WriteString(" → ")
				piece.WriteString(strconv.Itoa(len(diff.NewText)))
				piece.WriteString(" bytes)")
			}
			appendPiece(piece.String())
		default:
			// terminal blocks, image blocks, unknown future types —
			// ignore. We have no way to inline-render them.
		}
	}
	return b.String()
}

func (c *acpClient) handleUsageUpdate(data json.RawMessage) {
	var msg struct {
		Usage struct {
			InputTokens      int64 `json:"inputTokens"`
			OutputTokens     int64 `json:"outputTokens"`
			TotalTokens      int64 `json:"totalTokens"`
			CachedReadTokens int64 `json:"cachedReadTokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	c.usageMu.Lock()
	// Usage updates from ACP are cumulative snapshots, so take the latest.
	if msg.Usage.InputTokens > c.usage.InputTokens {
		c.usage.InputTokens = msg.Usage.InputTokens
	}
	if msg.Usage.OutputTokens > c.usage.OutputTokens {
		c.usage.OutputTokens = msg.Usage.OutputTokens
	}
	if msg.Usage.CachedReadTokens > c.usage.CacheReadTokens {
		c.usage.CacheReadTokens = msg.Usage.CachedReadTokens
	}
	c.usageMu.Unlock()
}

// ── Helpers ──

// extractACPSessionID pulls `sessionId` out of a session/new or
// session/resume response. Shared by all ACP backends (Kimi, kimi,
// and anything else that follows the standard ACP schema).
func extractACPSessionID(result json.RawMessage) string {
	var r struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.Unmarshal(result, &r); err != nil {
		return ""
	}
	return r.SessionID
}

// buildACPSessionParams constructs the params map for the ACP `session/new`
// request. The `model` field is only included when non-empty so Kimi falls
// back to its default only when no explicit model was configured.
func buildACPSessionParams(cwd, model string) map[string]any {
	params := map[string]any{
		"cwd":        cwd,
		"mcpServers": []any{},
	}
	if model != "" {
		params["model"] = model
	}
	return params
}

// acpToolNameFromTitle extracts a tool name from the ACP tool call title.
// Kimi ACP titles look like "terminal: ls -la", "read: /path/to/file", etc.
// Some titles have no colon (e.g. "execute code").
func acpToolNameFromTitle(title string, kind string) string {
	// Check exact-match titles first (no colon).
	switch title {
	case "execute code":
		return "execute_code"
	}

	// Try to extract the tool name from before the first colon.
	if idx := strings.Index(title, ":"); idx > 0 {
		name := strings.TrimSpace(title[:idx])
		// Map common ACP title prefixes back to tool names.
		// Some titles include mode info like "patch (replace)", so check prefix.
		switch {
		case name == "terminal":
			return "terminal"
		case name == "read":
			return "read_file"
		case name == "write":
			return "write_file"
		case strings.HasPrefix(name, "patch"):
			return "patch"
		case name == "search":
			return "search_files"
		case name == "web search":
			return "web_search"
		case name == "extract":
			return "web_extract"
		case name == "delegate":
			return "delegate_task"
		case name == "analyze image":
			return "vision_analyze"
		}
		return name
	}

	// Fall back to kind.
	switch kind {
	case "read":
		return "read_file"
	case "edit":
		return "write_file"
	case "execute":
		return "terminal"
	case "search":
		return "search_files"
	case "fetch":
		return "web_search"
	case "think":
		return "thinking"
	default:
		// Preserve a non-empty title when we can't classify it: kimi
		// emits bare titles like "Shell" or "Read file" without any
		// `kind`, so returning an empty string here drops the tool
		// name entirely before kimiToolNameFromTitle can map it.
		// Kimi titles always carry a colon, so Kimi never reaches
		// this branch with a non-empty title.
		if title != "" {
			return title
		}
		return kind
	}
}

// ── Provider-error sniffing ──
//
// ACP agents (Kimi, kimi, …) all have the same failure mode:
// session/prompt reports stopReason=end_turn even when the underlying
// HTTP call to the configured LLM endpoint returned an error — the
// actionable detail only appears on stderr (e.g.
// `⚠️ API call failed (attempt 1/3): BadRequestError [HTTP 400]` and
// `Error: HTTP 400: Error code: 400 - {'detail': "The '...' model
// is not supported when using Codex with a ChatGPT account."}`).
// The sniffer scans for those patterns so the daemon can surface a
// real failure instead of a generic "empty output".
//
// Parameterised by provider name so both Kimi and kimi can share
// the transport: the regexes match format-level signals (HTTP status,
// error-kind tags, "API call failed" banner) that both runtimes emit.
type acpProviderErrorSniffer struct {
	provider string
	mu       sync.Mutex
	remains  []byte   // buffer for a partial trailing line across writes
	lines    []string // captured error lines, bounded
	seen     map[string]bool
}

// acpErrorHeaderRe matches the first line of an API-error block.
// ACP agents typically prefix these with ⚠️ / ❌ and include an HTTP
// status code or a non-retryable-error tag.
var acpErrorHeaderRe = regexp.MustCompile(`(?:⚠️|❌|\[ERROR\]).*(?:BadRequestError|AuthenticationError|RateLimitError|HTTP [0-9]{3}|Non-retryable|API call failed)`)

// acpErrorDetailRe pulls the most useful single-line messages out of
// the subsequent lines of the error block (the one whose "Error:" or
// "Details:" tag actually spells out what happened).
var acpErrorDetailRe = regexp.MustCompile(`(?:Error:|detail:|Details:)\s*(.+)`)

const acpMaxErrorLines = 8

// newACPProviderErrorSniffer returns a sniffer that tags its messages
// with the given provider name (e.g. "Kimi", "kimi") so failure
// strings make it obvious which runtime produced the error.
func newACPProviderErrorSniffer(provider string) *acpProviderErrorSniffer {
	return &acpProviderErrorSniffer{provider: provider, seen: map[string]bool{}}
}

// Write implements io.Writer so the sniffer can sit behind an
// io.MultiWriter next to the normal stderr log forwarder.
func (s *acpProviderErrorSniffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := append(s.remains, p...)
	// Keep the final partial line (no trailing newline) for the
	// next write so multi-line error blocks aren't split.
	nl := strings.LastIndexByte(string(data), '\n')
	var complete string
	if nl < 0 {
		s.remains = append(s.remains[:0], data...)
		return len(p), nil
	}
	complete = string(data[:nl])
	s.remains = append(s.remains[:0], data[nl+1:]...)

	for _, line := range strings.Split(complete, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !(acpErrorHeaderRe.MatchString(line) || acpErrorDetailRe.MatchString(line)) {
			continue
		}
		if s.seen[line] {
			continue
		}
		s.seen[line] = true
		s.lines = append(s.lines, line)
		if len(s.lines) > acpMaxErrorLines {
			s.lines = s.lines[len(s.lines)-acpMaxErrorLines:]
		}
	}
	return len(p), nil
}

// message returns a single-line summary suitable for the task
// error field. Prefers the most specific "Error:" / "detail:"
// fragment; falls back to the first captured header line; empty
// when nothing useful was seen.
func (s *acpProviderErrorSniffer) message() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefix := s.provider + " provider error: "
	for _, line := range s.lines {
		if m := acpErrorDetailRe.FindStringSubmatch(line); m != nil {
			detail := strings.TrimSpace(m[1])
			if detail != "" {
				return prefix + detail
			}
		}
	}
	for _, line := range s.lines {
		if acpErrorHeaderRe.MatchString(line) {
			return prefix + line
		}
	}
	return ""
}
