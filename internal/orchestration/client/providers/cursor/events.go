package cursor

import (
	"encoding/json"

	"github.com/zjrosen/perles/internal/orchestration/client"
)

// cursorEvent represents the raw Cursor Agent CLI stream-json output structure.
// Cursor uses the same stream-json format as Claude Code, with events like:
//
//	{"type":"system","subtype":"init","session_id":"...","model":"..."}
//	{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"..."}]}}
//	{"type":"tool_use","tool":{"id":"...","name":"...","input":{...}}}
//	{"type":"result","result":"...","total_cost_usd":0.05,"duration_ms":1234}
type cursorEvent struct {
	Type          client.EventType             `json:"type"`
	SubType       string                       `json:"subtype,omitempty"`
	SessionID     string                       `json:"session_id,omitempty"`
	WorkDir       string                       `json:"cwd,omitempty"`
	Message       *cursorMessage               `json:"message,omitempty"`
	Tool          *client.ToolContent          `json:"tool,omitempty"`
	ModelUsage    map[string]client.ModelUsage `json:"modelUsage,omitempty"` //nolint:tagliatelle // stream-json uses camelCase
	Error         json.RawMessage              `json:"error,omitempty"`
	TotalCostUSD  float64                      `json:"total_cost_usd,omitempty"`
	DurationMs    int64                        `json:"duration_ms,omitempty"`
	IsErrorResult bool                         `json:"is_error,omitempty"`
	Result        string                       `json:"result,omitempty"`
}

// cursorMessage represents the message object in Cursor events.
type cursorMessage struct {
	ID         string              `json:"id,omitempty"`
	Role       string              `json:"role,omitempty"`
	Content    []cursorContentBlock `json:"content,omitempty"`
	Model      string              `json:"model,omitempty"`
	Usage      *cursorUsage        `json:"usage,omitempty"`
	StopReason string              `json:"stop_reason,omitempty"`
}

// cursorContentBlock represents a content block in Cursor messages.
type cursorContentBlock struct {
	Type  string          `json:"type,omitempty"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// cursorUsage holds token usage from Cursor CLI JSON output.
// Note: As of 2025, Cursor's stream-json output does not populate these fields.
// The struct is retained for forward-compatibility if Cursor adds usage reporting.
type cursorUsage struct {
	InputTokens              int `json:"input_tokens,omitempty"`
	OutputTokens             int `json:"output_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
}
