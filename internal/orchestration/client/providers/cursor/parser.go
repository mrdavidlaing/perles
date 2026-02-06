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
			event.Message.Content = append(event.Message.Content, client.ContentBlock{
				Type:  block.Type,
				Text:  block.Text,
				ID:    block.ID,
				Name:  block.Name,
				Input: block.Input,
			})
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
