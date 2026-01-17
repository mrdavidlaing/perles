package workflow

import "time"

// WorkflowState captures the active workflow state for persistence across
// coordinator refresh cycles.
type WorkflowState struct {
	WorkflowID      string    `json:"workflow_id"`      // Registry identifier
	WorkflowName    string    `json:"workflow_name"`    // Human-readable name
	WorkflowContent string    `json:"workflow_content"` // Full markdown content
	StartedAt       time.Time `json:"started_at"`       // When workflow was activated
}

// IsActive returns true if a workflow is currently active.
func (ws *WorkflowState) IsActive() bool {
	return ws != nil && ws.WorkflowID != ""
}

// WorkflowStateFilename is the filename used to persist workflow state within a session directory.
const WorkflowStateFilename = "workflow_state.json"
