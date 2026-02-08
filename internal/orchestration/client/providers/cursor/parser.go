package cursor

import (
	"encoding/json"
	"strings"

	"github.com/zjrosen/perles/internal/orchestration/client"
)

const (
	// CursorContextWindowSize is the assumed context window size for Cursor Agent.
	// Cursor supports multiple model providers with varying context windows;
	// 200000 is a conservative default. Cursor CLI does not report the actual
	// context window size, so this value is used for usage percentage estimates.
	CursorContextWindowSize = 200000

	// eventThinking is the Cursor-specific event type for model reasoning/thinking.
	// These events have subtypes "delta" (streaming chunks) and "completed".
	// They contain no user-visible content and are skipped during parsing.
	eventThinking client.EventType = "thinking"

	// eventToolCall is the Cursor-specific event type for tool invocations.
	// Unlike Claude's "tool_use" events, Cursor emits "tool_call" with subtypes
	// "started" and "completed", with a polymorphic tool_call body containing
	// one of: shellToolCall, mcpToolCall, editToolCall, readToolCall.
	eventToolCall client.EventType = "tool_call"
)

// Parser implements client.EventParser for Cursor Agent CLI stream-json events.
// Cursor uses the same stream-json format as Claude Code, so the parser
// follows the Claude parser structure closely.
type Parser struct {
	client.BaseParser
}

// NewParser creates a new Cursor EventParser with the default context window size.
func NewParser() *Parser {
	return &Parser{
		BaseParser: client.NewBaseParser(CursorContextWindowSize),
	}
}

// ParseEvent converts Cursor Agent CLI stream-json to client.OutputEvent.
// This is the main parsing entry point called for each stdout line.
func (p *Parser) ParseEvent(data []byte) (client.OutputEvent, error) {
	var raw cursorEvent
	if err := json.Unmarshal(data, &raw); err != nil {
		return client.OutputEvent{}, err
	}

	// Skip thinking events entirely. Cursor emits these for model reasoning
	// (subtypes: "delta" for streaming chunks, "completed" when done).
	// They contain no user-visible content and would pass through the V2
	// handler as no-ops, but skipping them avoids unnecessary event processing.
	if raw.Type == eventThinking {
		return client.OutputEvent{}, client.ErrSkipEvent
	}

	// Map Cursor tool_call events to the unified tool_use/tool_result types.
	// Cursor emits "tool_call" with subtypes "started" and "completed" instead of
	// Claude's "tool_use" and "tool_result" event types.
	if raw.Type == eventToolCall && raw.ToolCall != nil {
		return p.parseToolCallEvent(raw, data)
	}

	event := client.OutputEvent{
		Type:          raw.Type,
		SubType:       raw.SubType,
		SessionID:     raw.SessionID,
		WorkDir:       raw.WorkDir,
		Tool:          raw.Tool,
		ModelUsage:    raw.ModelUsage,
		TotalCostUSD:  raw.TotalCostUSD,
		DurationMs:    raw.DurationMs,
		IsErrorResult: raw.IsErrorResult,
		Result:        raw.Result,
	}

	// Parse error field - handle polymorphic error (string or object)
	event.Error = client.ParsePolymorphicError(raw.Error)

	if raw.Message != nil {
		event.Message = &client.MessageContent{
			ID:    raw.Message.ID,
			Role:  raw.Message.Role,
			Model: raw.Message.Model,
		}
		for _, block := range raw.Message.Content {
			text := block.Text
			// Cursor emits assistant text blocks with leading/trailing whitespace,
			// especially after thinking events (e.g., "\n\n" or "\nActual text\n").
			// Trim to prevent empty lines in the TUI output.
			if block.Type == "text" {
				text = strings.TrimSpace(text)
			}
			event.Message.Content = append(event.Message.Content, client.ContentBlock{
				Type:  block.Type,
				Text:  text,
				ID:    block.ID,
				Name:  block.Name,
				Input: block.Input,
			})
		}

		// If all text blocks are empty after trimming, skip the entire event.
		// This filters out whitespace-only assistant messages that Cursor emits
		// between thinking and substantive output.
		if event.Type == client.EventAssistant && event.Message.GetText() == "" && !event.Message.HasToolUses() {
			return client.OutputEvent{}, client.ErrSkipEvent
		}

		// Detect context exhaustion pattern:
		// - error code is "invalid_request"
		// - message content contains "Prompt is too long"
		if event.Error != nil && event.Error.Code == "invalid_request" {
			messageText := event.Message.GetText()
			if strings.Contains(messageText, "Prompt is too long") || raw.Message.StopReason == "stop_sequence" {
				event.Error.Reason = client.ErrReasonContextExceeded
				if event.Error.Message == "" {
					event.Error.Message = messageText
				}
			}
		}
	}

	// Extract token usage from assistant events
	if raw.Type == client.EventAssistant && raw.Message != nil && raw.Message.Usage != nil {
		tokensUsed := raw.Message.Usage.InputTokens + raw.Message.Usage.CacheReadInputTokens + raw.Message.Usage.CacheCreationInputTokens
		event.Usage = &client.UsageInfo{
			TokensUsed:   tokensUsed,
			TotalTokens:  p.ContextWindowSize(),
			OutputTokens: raw.Message.Usage.OutputTokens,
		}
	}

	// Copy raw data for debugging
	event.Raw = make([]byte, len(data))
	copy(event.Raw, data)

	return event, nil
}

// parseToolCallEvent converts Cursor's tool_call events to the unified OutputEvent format.
// Maps "started" subtype to EventToolUse and "completed" subtype to EventToolResult.
func (p *Parser) parseToolCallEvent(raw cursorEvent, data []byte) (client.OutputEvent, error) {
	tc := raw.ToolCall
	toolName := tc.toolName()

	event := client.OutputEvent{
		SessionID: raw.SessionID,
	}

	switch raw.SubType {
	case "started":
		// Map to tool_use event that the V2 handler understands.
		// The V2 process handler checks event.Message for tool_use content blocks,
		// so we wrap the tool info in a Message to match the expected structure.
		event.Type = client.EventToolUse
		event.Tool = &client.ToolContent{
			ID:    raw.CallID,
			Name:  toolName,
			Input: tc.toolInput(),
		}
		event.Message = &client.MessageContent{
			Role: "assistant",
			Content: []client.ContentBlock{
				{
					Type:  "tool_use",
					ID:    raw.CallID,
					Name:  toolName,
					Input: tc.toolInput(),
				},
			},
		}

	case "completed":
		// Map to tool_result event
		event.Type = client.EventToolResult
		output := tc.toolOutput()
		event.Tool = &client.ToolContent{
			ID:     raw.CallID,
			Name:   toolName,
			Output: output,
			Input:  tc.toolInput(),
		}

	default:
		// Unknown subtype, skip
		return client.OutputEvent{}, client.ErrSkipEvent
	}

	event.Raw = make([]byte, len(data))
	copy(event.Raw, data)

	return event, nil
}

// ExtractSessionRef returns the session identifier from an event.
// Cursor uses session_id in init events, similar to Claude Code.
func (p *Parser) ExtractSessionRef(_ client.OutputEvent, _ []byte) string {
	// Session extraction is handled via the extractSession function in process.go
	// using the OnInitEvent hook pattern, same as Claude.
	return ""
}

// IsContextExhausted checks if an event indicates context window exhaustion.
// This extends BaseParser's detection with Cursor-specific checks.
func (p *Parser) IsContextExhausted(event client.OutputEvent) bool {
	if p.BaseParser.IsContextExhausted(event) {
		return true
	}

	if event.Error != nil && event.Error.Reason == client.ErrReasonContextExceeded {
		return true
	}

	return false
}

// Verify Parser implements EventParser at compile time.
var _ client.EventParser = (*Parser)(nil)
