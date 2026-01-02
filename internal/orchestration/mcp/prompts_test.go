package mcp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// CommitApprovalPrompt Tests (Updated for post_accountability_summary)
// ============================================================================

// TestCommitApprovalPrompt_UsesAccountabilitySummary verifies post_accountability_summary is referenced.
func TestCommitApprovalPrompt_UsesAccountabilitySummary(t *testing.T) {
	taskID := "perles-abc.1"
	prompt := CommitApprovalPrompt(taskID, "")

	require.Contains(t, prompt, "post_accountability_summary",
		"Prompt should include post_accountability_summary instruction")
	require.NotContains(t, prompt, "post_reflections",
		"Prompt should NOT include post_reflections (deprecated)")
}

// TestCommitApprovalPrompt_IncludesAccountabilityFields verifies all new fields are documented.
func TestCommitApprovalPrompt_IncludesAccountabilityFields(t *testing.T) {
	prompt := CommitApprovalPrompt("test-task", "")

	fields := []string{
		"task_id",
		"summary",
		"commits",
		"issues_closed",
		"issues_discovered",
		"verification_points",
		"retro",
		"next_steps",
	}

	for _, field := range fields {
		require.Contains(t, prompt, field,
			"Prompt should document field: %s", field)
	}
}

// TestCommitApprovalPrompt_IncludesRetroStructure verifies retro feedback structure.
func TestCommitApprovalPrompt_IncludesRetroStructure(t *testing.T) {
	prompt := CommitApprovalPrompt("test-task", "")

	// The example should show the retro structure
	require.Contains(t, prompt, "went_well", "Prompt should show went_well in retro")
	require.Contains(t, prompt, "friction", "Prompt should show friction in retro")
	require.Contains(t, prompt, "patterns", "Prompt should show patterns in retro")
}
