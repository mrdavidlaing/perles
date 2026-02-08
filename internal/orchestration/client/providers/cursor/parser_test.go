package cursor

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zjrosen/perles/internal/orchestration/client"
)

func TestParser_NewParser(t *testing.T) {
	p := NewParser()
	require.NotNil(t, p, "NewParser should return non-nil parser")
}

func TestParser_ContextWindowSize(t *testing.T) {
	p := NewParser()
	size := p.ContextWindowSize()
	require.Equal(t, 200000, size, "ContextWindowSize should return 200000 for Cursor")
}

func TestParser_ParseEvent_AssistantMessage(t *testing.T) {
	p := NewParser()

	// Cursor stream-json assistant event (same format as Claude)
	input := `{"type":"assistant","message":{"id":"msg_123","role":"assistant","content":[{"type":"text","text":"Hello, I can help with that."}],"model":"composer-1"}}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventAssistant, event.Type)
	require.NotNil(t, event.Message)
	require.Equal(t, "assistant", event.Message.Role)
	require.Equal(t, "composer-1", event.Message.Model)
	require.Len(t, event.Message.Content, 1)
	require.Equal(t, "text", event.Message.Content[0].Type)
	require.Equal(t, "Hello, I can help with that.", event.Message.Content[0].Text)
}

func TestParser_ParseEvent_SystemInit(t *testing.T) {
	p := NewParser()

	input := `{"type":"system","subtype":"init","session_id":"ses_cursor_123","cwd":"/project"}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventSystem, event.Type)
	require.Equal(t, "init", event.SubType)
	require.Equal(t, "ses_cursor_123", event.SessionID)
	require.Equal(t, "/project", event.WorkDir)
}

func TestParser_ParseEvent_ToolUse(t *testing.T) {
	p := NewParser()

	input := `{"type":"tool_use","tool":{"id":"tool_123","name":"Bash","input":{"command":"ls -la"}}}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventToolUse, event.Type)
	require.NotNil(t, event.Tool)
	require.Equal(t, "tool_123", event.Tool.ID)
	require.Equal(t, "Bash", event.Tool.Name)
}

func TestParser_ParseEvent_Result(t *testing.T) {
	p := NewParser()

	input := `{"type":"result","result":"Task completed successfully","total_cost_usd":0.0523,"duration_ms":45000}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventResult, event.Type)
	require.Equal(t, "Task completed successfully", event.Result)
	require.InDelta(t, 0.0523, event.TotalCostUSD, 0.0001)
	require.Equal(t, int64(45000), event.DurationMs)
}

func TestParser_ParseEvent_ErrorAsString(t *testing.T) {
	p := NewParser()

	// When error is a plain string, ParsePolymorphicError puts it in Message field
	input := `{"type":"assistant","error":"invalid_request","message":{"role":"assistant","content":[{"type":"text","text":"test"}]}}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.NotNil(t, event.Error)
	require.Equal(t, "invalid_request", event.Error.Message)
}

func TestParser_ParseEvent_ErrorAsObject(t *testing.T) {
	p := NewParser()

	input := `{"type":"assistant","error":{"code":"rate_limit","message":"Rate limit exceeded"}}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.NotNil(t, event.Error)
	require.Equal(t, "rate_limit", event.Error.Code)
	require.Equal(t, "Rate limit exceeded", event.Error.Message)
}

func TestParser_ParseEvent_WithUsage(t *testing.T) {
	p := NewParser()

	input := `{"type":"assistant","message":{"id":"msg_456","role":"assistant","content":[{"type":"text","text":"done"}],"usage":{"input_tokens":5000,"output_tokens":1000,"cache_read_input_tokens":2000,"cache_creation_input_tokens":500}}}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventAssistant, event.Type)
	require.NotNil(t, event.Usage)
	require.Equal(t, 7500, event.Usage.TokensUsed) // input + cache_read + cache_creation
	require.Equal(t, 200000, event.Usage.TotalTokens)
	require.Equal(t, 1000, event.Usage.OutputTokens)
}

func TestParser_ParseEvent_InvalidJSON(t *testing.T) {
	p := NewParser()

	_, err := p.ParseEvent([]byte("not json"))
	require.Error(t, err)
}

func TestParser_ParseEvent_EmptyObject(t *testing.T) {
	p := NewParser()

	event, err := p.ParseEvent([]byte("{}"))
	require.NoError(t, err)
	require.Equal(t, client.EventType(""), event.Type)
}

func TestParser_ParseEvent_ContextExhausted(t *testing.T) {
	p := NewParser()

	// Error event with context exhaustion pattern using structured error object
	input := `{"type":"assistant","error":{"code":"invalid_request","message":"Prompt is too long: 250000 tokens"},"message":{"role":"assistant","content":[{"type":"text","text":"Prompt is too long: 250000 tokens"}],"stop_reason":"stop_sequence"}}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.NotNil(t, event.Error)
	require.Equal(t, "invalid_request", event.Error.Code)
	require.Equal(t, client.ErrReasonContextExceeded, event.Error.Reason)
}

func TestParser_ExtractSessionRef_ReturnsEmpty(t *testing.T) {
	p := NewParser()

	// Cursor uses the OnInitEvent hook pattern (like Claude), so ExtractSessionRef returns empty
	result := p.ExtractSessionRef(client.OutputEvent{}, nil)
	require.Equal(t, "", result)
}

func TestParser_IsContextExhausted(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name  string
		event client.OutputEvent
		want  bool
	}{
		{
			name:  "empty event not exhausted",
			event: client.OutputEvent{},
			want:  false,
		},
		{
			name: "error without context pattern",
			event: client.OutputEvent{
				Type:  client.EventError,
				Error: &client.ErrorInfo{Message: "Connection refused"},
			},
			want: false,
		},
		{
			name: "error with ErrReasonContextExceeded",
			event: client.OutputEvent{
				Type:  client.EventError,
				Error: &client.ErrorInfo{Message: "error", Reason: client.ErrReasonContextExceeded},
			},
			want: true,
		},
		{
			name: "error with prompt too long pattern",
			event: client.OutputEvent{
				Type:  client.EventError,
				Error: &client.ErrorInfo{Message: "Prompt is too long: 300000 tokens"},
			},
			want: true,
		},
		{
			name: "error with context window exceeded pattern",
			event: client.OutputEvent{
				Type:  client.EventError,
				Error: &client.ErrorInfo{Message: "Context window exceeded"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.IsContextExhausted(tt.event)
			require.Equal(t, tt.want, result)
		})
	}
}

func TestParser_ImplementsEventParser(t *testing.T) {
	p := NewParser()

	// Verify at runtime that Parser implements EventParser
	var _ client.EventParser = p
}

func TestParser_ParseEvent_ModelUsage(t *testing.T) {
	p := NewParser()

	input := `{"type":"result","result":"done","modelUsage":{"composer-1":{"inputTokens":5000,"outputTokens":1000,"cacheReadInputTokens":2000,"costUSD":0.05}}}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventResult, event.Type)
	require.NotNil(t, event.ModelUsage)
	require.Contains(t, event.ModelUsage, "composer-1")
	require.Equal(t, 5000, event.ModelUsage["composer-1"].InputTokens)
	require.Equal(t, 1000, event.ModelUsage["composer-1"].OutputTokens)
}

func TestParser_ParseEvent_IsErrorResult(t *testing.T) {
	p := NewParser()

	input := `{"type":"result","is_error":true,"result":"Something went wrong"}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventResult, event.Type)
	require.True(t, event.IsErrorResult)
	require.Equal(t, "Something went wrong", event.Result)
	require.True(t, event.IsError())
}

func TestParser_ParseEvent_ThinkingDelta_Skipped(t *testing.T) {
	p := NewParser()

	input := `{"type":"thinking","subtype":"delta","text":"","session_id":"3c5ee206-ba17-4781-927c-3b5a53a4e7a8","timestamp_ms":1770503338548}`

	_, err := p.ParseEvent([]byte(input))

	require.ErrorIs(t, err, client.ErrSkipEvent)
}

func TestParser_ParseEvent_ThinkingCompleted_Skipped(t *testing.T) {
	p := NewParser()

	input := `{"type":"thinking","subtype":"completed","session_id":"174e78e5-d274-4a42-a78d-41740cab171c","timestamp_ms":1770503339073}`

	_, err := p.ParseEvent([]byte(input))

	require.ErrorIs(t, err, client.ErrSkipEvent)
}

func TestParser_ParseEvent_WhitespaceOnlyAssistant_Skipped(t *testing.T) {
	p := NewParser()

	// After thinking completes, Cursor often emits assistant messages with only newlines
	input := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"\n\n"}]},"session_id":"3c5ee206"}`

	_, err := p.ParseEvent([]byte(input))

	require.ErrorIs(t, err, client.ErrSkipEvent)
}

func TestParser_ParseEvent_AssistantTextTrimmed(t *testing.T) {
	p := NewParser()

	// Cursor wraps assistant text in leading/trailing newlines
	input := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"\nReading the epic to understand the workflow.\n"}]},"session_id":"174e78e5"}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventAssistant, event.Type)
	require.Equal(t, "Reading the epic to understand the workflow.", event.Message.GetText())
}

func TestParser_ParseEvent_AssistantWithToolUse_NotSkipped(t *testing.T) {
	p := NewParser()

	// Even if text blocks are empty, tool_use blocks should preserve the event
	input := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"\n"},{"type":"tool_use","id":"tool_1","name":"Bash","input":{"command":"ls"}}]},"session_id":"abc123"}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventAssistant, event.Type)
	require.True(t, event.Message.HasToolUses())
}

func TestParser_ParseEvent_ToolCallStarted_Shell(t *testing.T) {
	p := NewParser()

	input := `{"type":"tool_call","subtype":"started","call_id":"tool_abc123","tool_call":{"shellToolCall":{"args":{"command":"bd show perles-lu5d --json"}}},"session_id":"ses_123"}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventToolUse, event.Type)
	require.Equal(t, "ses_123", event.SessionID)
	require.NotNil(t, event.Tool)
	require.Equal(t, "tool_abc123", event.Tool.ID)
	require.Equal(t, "Bash", event.Tool.Name)
	// Should also set Message with tool_use content block for V2 handler
	require.NotNil(t, event.Message)
	require.True(t, event.Message.HasToolUses())
	require.Equal(t, "Bash", event.Message.Content[0].Name)
}

func TestParser_ParseEvent_ToolCallCompleted_Shell(t *testing.T) {
	p := NewParser()

	input := `{"type":"tool_call","subtype":"completed","call_id":"tool_abc123","tool_call":{"shellToolCall":{"args":{"command":"echo hello"},"result":{"success":{"stdout":"hello\n","stderr":"","exitCode":0}}}},"session_id":"ses_123"}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventToolResult, event.Type)
	require.NotNil(t, event.Tool)
	require.Equal(t, "tool_abc123", event.Tool.ID)
	require.Equal(t, "Bash", event.Tool.Name)
	require.Equal(t, "hello\n", event.Tool.GetOutput())
}

func TestParser_ParseEvent_ToolCallStarted_MCP(t *testing.T) {
	p := NewParser()

	input := `{"type":"tool_call","subtype":"started","call_id":"tool_mcp_456","tool_call":{"mcpToolCall":{"args":{"name":"perles-orchestrator-spawn_worker","args":{"agent_type":"implementer"},"toolName":"spawn_worker"}}},"session_id":"ses_456"}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventToolUse, event.Type)
	require.NotNil(t, event.Tool)
	require.Equal(t, "spawn_worker", event.Tool.Name)
	require.NotNil(t, event.Message)
	require.Equal(t, "spawn_worker", event.Message.Content[0].Name)
}

func TestParser_ParseEvent_ToolCallCompleted_MCP(t *testing.T) {
	p := NewParser()

	input := `{"type":"tool_call","subtype":"completed","call_id":"tool_mcp_456","tool_call":{"mcpToolCall":{"args":{"name":"perles-orchestrator-spawn_worker","args":{"agent_type":"implementer"},"toolName":"spawn_worker"},"result":{"success":{"content":[{"text":{"text":"Process worker-1 spawned"}}],"isError":false}}}},"session_id":"ses_456"}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventToolResult, event.Type)
	require.NotNil(t, event.Tool)
	require.Equal(t, "spawn_worker", event.Tool.Name)
	require.Equal(t, "Process worker-1 spawned", event.Tool.GetOutput())
}

func TestParser_ParseEvent_ToolCallStarted_Edit(t *testing.T) {
	p := NewParser()

	input := `{"type":"tool_call","subtype":"started","call_id":"tool_edit_789","tool_call":{"editToolCall":{"args":{"path":"/tmp/test.md","streamContent":"# Test"}}},"session_id":"ses_789"}`

	event, err := p.ParseEvent([]byte(input))

	require.NoError(t, err)
	require.Equal(t, client.EventToolUse, event.Type)
	require.NotNil(t, event.Tool)
	require.Equal(t, "Edit", event.Tool.Name)
}

func TestParser_ParseEvent_ToolCallUnknownSubtype_Skipped(t *testing.T) {
	p := NewParser()

	input := `{"type":"tool_call","subtype":"cancelled","call_id":"tool_xxx","tool_call":{"shellToolCall":{"args":{"command":"test"}}}}`

	_, err := p.ParseEvent([]byte(input))

	require.ErrorIs(t, err, client.ErrSkipEvent)
}
