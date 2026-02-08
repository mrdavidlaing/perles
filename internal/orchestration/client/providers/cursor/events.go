package cursor

import (
	"encoding/json"

	"github.com/zjrosen/perles/internal/orchestration/client"
)

// cursorEvent represents the raw Cursor Agent CLI stream-json output structure.
// Cursor emits events in two formats:
//
// Claude-compatible format:
//
//	{"type":"system","subtype":"init","session_id":"...","model":"..."}
//	{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"..."}]}}
//	{"type":"result","result":"...","total_cost_usd":0.05,"duration_ms":1234}
//
// Cursor-specific format (tool calls):
//
//	{"type":"tool_call","subtype":"started","call_id":"...","tool_call":{"shellToolCall":{"args":{...}}}}
//	{"type":"tool_call","subtype":"completed","call_id":"...","tool_call":{"mcpToolCall":{...,"result":{...}}}}
//	{"type":"thinking","subtype":"delta","text":"..."}
//	{"type":"thinking","subtype":"completed"}
type cursorEvent struct {
	Type          client.EventType             `json:"type"`
	SubType       string                       `json:"subtype,omitempty"`
	SessionID     string                       `json:"session_id,omitempty"`
	WorkDir       string                       `json:"cwd,omitempty"`
	Message       *cursorMessage               `json:"message,omitempty"`
	Tool          *client.ToolContent          `json:"tool,omitempty"`
	ToolCall      *cursorToolCall              `json:"tool_call,omitempty"`  //nolint:tagliatelle // Cursor uses snake_case
	CallID        string                       `json:"call_id,omitempty"`    //nolint:tagliatelle // Cursor uses snake_case
	ModelUsage    map[string]client.ModelUsage `json:"modelUsage,omitempty"` //nolint:tagliatelle // stream-json uses camelCase
	Error         json.RawMessage              `json:"error,omitempty"`
	TotalCostUSD  float64                      `json:"total_cost_usd,omitempty"`
	DurationMs    int64                        `json:"duration_ms,omitempty"`
	IsErrorResult bool                         `json:"is_error,omitempty"`
	Result        string                       `json:"result,omitempty"`
}

// cursorToolCall represents the polymorphic tool_call object in Cursor events.
// Exactly one of the fields will be non-nil depending on the tool type.
type cursorToolCall struct {
	ShellToolCall *cursorShellToolCall `json:"shellToolCall,omitempty"` //nolint:tagliatelle // Cursor uses camelCase
	MCPToolCall   *cursorMCPToolCall   `json:"mcpToolCall,omitempty"`   //nolint:tagliatelle // Cursor uses camelCase
	EditToolCall  *cursorEditToolCall  `json:"editToolCall,omitempty"`  //nolint:tagliatelle // Cursor uses camelCase
	ReadToolCall  *cursorReadToolCall  `json:"readToolCall,omitempty"`  //nolint:tagliatelle // Cursor uses camelCase
}

// toolName returns the display name for this tool call.
func (tc *cursorToolCall) toolName() string {
	switch {
	case tc.ShellToolCall != nil:
		return "Bash"
	case tc.MCPToolCall != nil:
		if tc.MCPToolCall.Args.ToolName != "" {
			return tc.MCPToolCall.Args.ToolName
		}
		return tc.MCPToolCall.Args.Name
	case tc.EditToolCall != nil:
		return "Edit"
	case tc.ReadToolCall != nil:
		return "Read"
	default:
		return "unknown"
	}
}

// toolInput returns a JSON representation of the tool input for display.
func (tc *cursorToolCall) toolInput() json.RawMessage {
	switch {
	case tc.ShellToolCall != nil:
		data, _ := json.Marshal(tc.ShellToolCall.Args)
		return data
	case tc.MCPToolCall != nil:
		data, _ := json.Marshal(tc.MCPToolCall.Args.Args)
		return data
	case tc.EditToolCall != nil:
		data, _ := json.Marshal(tc.EditToolCall.Args)
		return data
	case tc.ReadToolCall != nil:
		data, _ := json.Marshal(tc.ReadToolCall.Args)
		return data
	default:
		return nil
	}
}

// toolOutput returns the result text from a completed tool call.
func (tc *cursorToolCall) toolOutput() string {
	switch {
	case tc.ShellToolCall != nil && tc.ShellToolCall.Result != nil:
		if s := tc.ShellToolCall.Result.Success; s != nil {
			return s.Stdout
		}
	case tc.MCPToolCall != nil && tc.MCPToolCall.Result != nil:
		if s := tc.MCPToolCall.Result.Success; s != nil {
			for _, c := range s.Content {
				if c.Text.Text != "" {
					return c.Text.Text
				}
			}
		}
	case tc.EditToolCall != nil && tc.EditToolCall.Result != nil:
		if s := tc.EditToolCall.Result.Success; s != nil {
			return s.Message
		}
	case tc.ReadToolCall != nil && tc.ReadToolCall.Result != nil:
		if s := tc.ReadToolCall.Result.Error; s != nil {
			return s.ErrorMessage
		}
	}
	return ""
}

// cursorShellToolCall represents a shell/bash tool call.
type cursorShellToolCall struct {
	Args   cursorShellArgs    `json:"args"`
	Result *cursorShellResult `json:"result,omitempty"`
}

type cursorShellArgs struct {
	Command string `json:"command,omitempty"`
}

type cursorShellResult struct {
	Success *cursorShellSuccess `json:"success,omitempty"`
}

type cursorShellSuccess struct {
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exitCode,omitempty"` //nolint:tagliatelle // Cursor uses camelCase
}

// cursorMCPToolCall represents an MCP tool call.
type cursorMCPToolCall struct {
	Args   cursorMCPArgs    `json:"args"`
	Result *cursorMCPResult `json:"result,omitempty"`
}

type cursorMCPArgs struct {
	Name     string          `json:"name,omitempty"`
	Args     json.RawMessage `json:"args,omitempty"`
	ToolName string          `json:"toolName,omitempty"` //nolint:tagliatelle // Cursor uses camelCase
}

type cursorMCPResult struct {
	Success *cursorMCPSuccess `json:"success,omitempty"`
}

type cursorMCPSuccess struct {
	Content []cursorMCPContent `json:"content,omitempty"`
	IsError bool               `json:"isError,omitempty"` //nolint:tagliatelle // Cursor uses camelCase
}

type cursorMCPContent struct {
	Text struct {
		Text string `json:"text,omitempty"`
	} `json:"text"`
}

// cursorEditToolCall represents a file edit tool call.
type cursorEditToolCall struct {
	Args   cursorEditArgs    `json:"args"`
	Result *cursorEditResult `json:"result,omitempty"`
}

type cursorEditArgs struct {
	Path          string `json:"path,omitempty"`
	StreamContent string `json:"streamContent,omitempty"` //nolint:tagliatelle // Cursor uses camelCase
}

type cursorEditResult struct {
	Success *cursorEditSuccess `json:"success,omitempty"`
}

type cursorEditSuccess struct {
	Path    string `json:"path,omitempty"`
	Message string `json:"message,omitempty"`
}

// cursorReadToolCall represents a file read tool call.
type cursorReadToolCall struct {
	Args   cursorReadArgs    `json:"args"`
	Result *cursorReadResult `json:"result,omitempty"`
}

type cursorReadArgs struct {
	Path string `json:"path,omitempty"`
}

type cursorReadResult struct {
	Error *cursorReadError `json:"error,omitempty"`
}

type cursorReadError struct {
	ErrorMessage string `json:"errorMessage,omitempty"` //nolint:tagliatelle // Cursor uses camelCase
}

// cursorMessage represents the message object in Cursor events.
type cursorMessage struct {
	ID         string               `json:"id,omitempty"`
	Role       string               `json:"role,omitempty"`
	Content    []cursorContentBlock `json:"content,omitempty"`
	Model      string               `json:"model,omitempty"`
	Usage      *cursorUsage         `json:"usage,omitempty"`
	StopReason string               `json:"stop_reason,omitempty"`
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
