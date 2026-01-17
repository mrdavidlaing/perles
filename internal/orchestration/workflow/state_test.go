package workflow

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowState_Creation(t *testing.T) {
	now := time.Now()
	state := WorkflowState{
		WorkflowID:      "debate",
		WorkflowName:    "Technical Debate",
		WorkflowContent: "# Debate Workflow\n\nSome content...",
		StartedAt:       now,
	}

	assert.Equal(t, "debate", state.WorkflowID)
	assert.Equal(t, "Technical Debate", state.WorkflowName)
	assert.Equal(t, "# Debate Workflow\n\nSome content...", state.WorkflowContent)
	assert.Equal(t, now, state.StartedAt)
}

func TestWorkflowState_IsActive(t *testing.T) {
	t.Run("returns true for valid workflow", func(t *testing.T) {
		state := &WorkflowState{
			WorkflowID:   "debate",
			WorkflowName: "Technical Debate",
		}
		assert.True(t, state.IsActive())
	})

	t.Run("returns false for nil workflow", func(t *testing.T) {
		var state *WorkflowState
		assert.False(t, state.IsActive())
	})

	t.Run("returns false for empty workflow ID", func(t *testing.T) {
		state := &WorkflowState{
			WorkflowID:   "",
			WorkflowName: "Technical Debate",
		}
		assert.False(t, state.IsActive())
	})

	t.Run("returns true with only ID set", func(t *testing.T) {
		state := &WorkflowState{
			WorkflowID: "debate",
		}
		assert.True(t, state.IsActive())
	})
}

func TestWorkflowState_JSONMarshalingRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second) // Truncate for JSON precision
	original := WorkflowState{
		WorkflowID:      "debate",
		WorkflowName:    "Technical Debate",
		WorkflowContent: "# Debate Workflow\n\nSome **markdown** content",
		StartedAt:       now,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var restored WorkflowState
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Verify all fields preserved
	assert.Equal(t, original.WorkflowID, restored.WorkflowID)
	assert.Equal(t, original.WorkflowName, restored.WorkflowName)
	assert.Equal(t, original.WorkflowContent, restored.WorkflowContent)
	assert.True(t, original.StartedAt.Equal(restored.StartedAt))
}

func TestWorkflowState_JSONMarshalingFormat(t *testing.T) {
	state := WorkflowState{
		WorkflowID:      "test-workflow",
		WorkflowName:    "Test Workflow",
		WorkflowContent: "Content here",
		StartedAt:       time.Date(2026, 1, 17, 12, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(state)
	require.NoError(t, err)

	// Verify JSON keys match the expected snake_case format
	var rawMap map[string]interface{}
	err = json.Unmarshal(data, &rawMap)
	require.NoError(t, err)

	assert.Contains(t, rawMap, "workflow_id")
	assert.Contains(t, rawMap, "workflow_name")
	assert.Contains(t, rawMap, "workflow_content")
	assert.Contains(t, rawMap, "started_at")
}

func TestWorkflowStateFilename(t *testing.T) {
	assert.Equal(t, "workflow_state.json", WorkflowStateFilename)
}
