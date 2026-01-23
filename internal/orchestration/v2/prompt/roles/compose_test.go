package roles

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// ComposeSystemPrompt Tests
// ============================================================================

// TestComposeSystemPrompt_BaseOnly verifies nil config returns base prompt.
func TestComposeSystemPrompt_BaseOnly(t *testing.T) {
	workerID := "worker-42"

	// Get expected base prompt
	rolePrompts := GetPrompts(AgentTypeGeneric)
	expectedBase := rolePrompts.SystemPrompt(workerID)

	// Compose with nil config
	result := ComposeSystemPrompt(workerID, AgentTypeGeneric, nil)

	require.Equal(t, expectedBase, result,
		"Nil config should return base prompt")
	require.Contains(t, result, workerID,
		"Result should contain workerID")
}

// TestComposeSystemPrompt_AppendMode verifies append is concatenated to base.
func TestComposeSystemPrompt_AppendMode(t *testing.T) {
	workerID := "worker-append-test"
	appendText := "## Workflow-specific instructions\n\nAlways use tabs for indentation."

	config := &WorkflowConfig{
		SystemPromptAppend: appendText,
	}

	// Get expected base prompt
	rolePrompts := GetPrompts(AgentTypeImplementer)
	expectedBase := rolePrompts.SystemPrompt(workerID)

	// Compose with append config
	result := ComposeSystemPrompt(workerID, AgentTypeImplementer, config)

	require.Contains(t, result, expectedBase,
		"Result should contain full base prompt")
	require.Contains(t, result, appendText,
		"Result should contain append text")
	require.True(t, strings.HasSuffix(result, appendText),
		"Append text should be at the end")
	require.Equal(t, expectedBase+"\n\n"+appendText, result,
		"Result should be base + newlines + append")
}

// TestComposeSystemPrompt_OverrideMode verifies override replaces base entirely.
func TestComposeSystemPrompt_OverrideMode(t *testing.T) {
	workerID := "worker-override-test"
	overrideText := "You are a completely custom agent with unique instructions."

	config := &WorkflowConfig{
		SystemPromptOverride: overrideText,
	}

	// Get base prompt (should NOT appear in result)
	rolePrompts := GetPrompts(AgentTypeReviewer)
	basePrompt := rolePrompts.SystemPrompt(workerID)

	// Compose with override config
	result := ComposeSystemPrompt(workerID, AgentTypeReviewer, config)

	require.Equal(t, overrideText, result,
		"Override should completely replace base prompt")
	require.NotContains(t, result, basePrompt[:50],
		"Result should not contain base prompt content")
}

// TestComposeSystemPrompt_OverrideEmptyString verifies empty override uses append/base.
func TestComposeSystemPrompt_OverrideEmptyString(t *testing.T) {
	workerID := "worker-empty-override"
	appendText := "This should be used when override is empty."

	config := &WorkflowConfig{
		SystemPromptOverride: "", // Empty string
		SystemPromptAppend:   appendText,
	}

	// Get expected base prompt
	rolePrompts := GetPrompts(AgentTypeResearcher)
	expectedBase := rolePrompts.SystemPrompt(workerID)

	// Compose - should fall through to append mode
	result := ComposeSystemPrompt(workerID, AgentTypeResearcher, config)

	require.Equal(t, expectedBase+"\n\n"+appendText, result,
		"Empty override should fall through to append mode")
}

// TestComposeSystemPrompt_AppendPreservesNewlines verifies proper formatting.
func TestComposeSystemPrompt_AppendPreservesNewlines(t *testing.T) {
	workerID := "worker-newlines"
	appendText := "Line 1\n\nLine 2\n\nLine 3"

	config := &WorkflowConfig{
		SystemPromptAppend: appendText,
	}

	result := ComposeSystemPrompt(workerID, AgentTypeGeneric, config)

	// Verify double newline separator between base and append
	require.Contains(t, result, "\n\n"+appendText,
		"Should have double newline before append content")

	// Verify newlines within append are preserved
	require.Contains(t, result, "Line 1\n\nLine 2",
		"Internal newlines in append should be preserved")
}

// TestComposeSystemPrompt_AllAgentTypes verifies works for all agent types.
func TestComposeSystemPrompt_AllAgentTypes(t *testing.T) {
	agentTypes := []AgentType{
		AgentTypeGeneric,
		AgentTypeImplementer,
		AgentTypeReviewer,
		AgentTypeResearcher,
	}

	workerID := "worker-all-types"
	appendText := "Universal workflow instruction."

	for _, agentType := range agentTypes {
		t.Run(agentType.String(), func(t *testing.T) {
			// Test nil config
			resultNil := ComposeSystemPrompt(workerID, agentType, nil)
			require.NotEmpty(t, resultNil, "Nil config should return non-empty prompt")
			require.Contains(t, resultNil, workerID, "Should contain workerID")

			// Test append config
			config := &WorkflowConfig{
				SystemPromptAppend: appendText,
			}
			resultAppend := ComposeSystemPrompt(workerID, agentType, config)
			require.Contains(t, resultAppend, appendText, "Should contain append text")
			require.Contains(t, resultAppend, resultNil[:50], "Should contain base prompt prefix")
		})
	}
}

// TestComposeSystemPrompt_OverrideTakesPrecedenceOverAppend verifies override wins.
func TestComposeSystemPrompt_OverrideTakesPrecedenceOverAppend(t *testing.T) {
	workerID := "worker-precedence"
	overrideText := "Override wins."
	appendText := "Append should be ignored."

	config := &WorkflowConfig{
		SystemPromptOverride: overrideText,
		SystemPromptAppend:   appendText,
	}

	result := ComposeSystemPrompt(workerID, AgentTypeGeneric, config)

	require.Equal(t, overrideText, result,
		"Override should take precedence over append")
	require.NotContains(t, result, appendText,
		"Append should not appear when override is set")
}

// TestComposeSystemPrompt_EmptyConfig verifies empty config returns base.
func TestComposeSystemPrompt_EmptyConfig(t *testing.T) {
	workerID := "worker-empty-config"
	config := &WorkflowConfig{
		// All fields empty/nil
	}

	rolePrompts := GetPrompts(AgentTypeImplementer)
	expectedBase := rolePrompts.SystemPrompt(workerID)

	result := ComposeSystemPrompt(workerID, AgentTypeImplementer, config)

	require.Equal(t, expectedBase, result,
		"Empty config should return base prompt")
}

// ============================================================================
// ComposeInitialPrompt Tests
// ============================================================================

// TestComposeInitialPrompt_BaseOnly verifies nil config returns base prompt.
func TestComposeInitialPrompt_BaseOnly(t *testing.T) {
	workerID := "worker-initial-42"

	// Get expected base prompt
	rolePrompts := GetPrompts(AgentTypeGeneric)
	expectedBase := rolePrompts.InitialPrompt(workerID)

	// Compose with nil config
	result := ComposeInitialPrompt(workerID, AgentTypeGeneric, nil)

	require.Equal(t, expectedBase, result,
		"Nil config should return base initial prompt")
	require.Contains(t, result, workerID,
		"Result should contain workerID")
}

// TestComposeInitialPrompt_AppendMode verifies append is added to base prompt.
func TestComposeInitialPrompt_AppendMode(t *testing.T) {
	workerID := "worker-initial-append"
	appendText := "Additional initial instructions for this workflow."

	config := &WorkflowConfig{
		InitialPromptAppend: appendText,
	}

	// Get expected base prompt
	rolePrompts := GetPrompts(AgentTypeReviewer)
	expectedBase := rolePrompts.InitialPrompt(workerID)

	result := ComposeInitialPrompt(workerID, AgentTypeReviewer, config)

	require.Contains(t, result, expectedBase,
		"Result should contain base initial prompt")
	require.Contains(t, result, appendText,
		"Result should contain appended text")
	require.True(t, strings.HasPrefix(result, expectedBase),
		"Base prompt should come first")
}

// TestComposeInitialPrompt_OverrideMode verifies override replaces base prompt.
func TestComposeInitialPrompt_OverrideMode(t *testing.T) {
	workerID := "worker-initial-override"
	overrideText := "Completely custom initial prompt for this workflow."

	config := &WorkflowConfig{
		InitialPromptOverride: overrideText,
	}

	result := ComposeInitialPrompt(workerID, AgentTypeGeneric, config)

	require.Equal(t, overrideText, result,
		"Override should completely replace base prompt")
}

// TestComposeInitialPrompt_OverrideTakesPrecedence verifies override wins over append.
func TestComposeInitialPrompt_OverrideTakesPrecedence(t *testing.T) {
	workerID := "worker-initial-precedence"
	overrideText := "Override wins."
	appendText := "This should be ignored."

	config := &WorkflowConfig{
		InitialPromptOverride: overrideText,
		InitialPromptAppend:   appendText,
	}

	result := ComposeInitialPrompt(workerID, AgentTypeImplementer, config)

	require.Equal(t, overrideText, result,
		"Override should take precedence over append")
	require.NotContains(t, result, appendText,
		"Append should not appear when override is set")
}

// TestComposeInitialPrompt_AllAgentTypes verifies works for all agent types.
func TestComposeInitialPrompt_AllAgentTypes(t *testing.T) {
	agentTypes := []AgentType{
		AgentTypeGeneric,
		AgentTypeImplementer,
		AgentTypeReviewer,
		AgentTypeResearcher,
	}

	workerID := "worker-initial-all"
	appendText := "Custom workflow instructions."

	for _, agentType := range agentTypes {
		t.Run(agentType.String(), func(t *testing.T) {
			// Test nil config returns base
			resultNil := ComposeInitialPrompt(workerID, agentType, nil)
			require.NotEmpty(t, resultNil, "Should return non-empty prompt")
			require.Contains(t, resultNil, workerID, "Should contain workerID")

			// Test with InitialPromptAppend
			config := &WorkflowConfig{
				InitialPromptAppend: appendText,
			}
			resultConfig := ComposeInitialPrompt(workerID, agentType, config)
			require.Contains(t, resultConfig, resultNil,
				"Should contain base prompt")
			require.Contains(t, resultConfig, appendText,
				"Should contain appended text")
		})
	}
}

// TestComposeInitialPrompt_EmptyConfig verifies empty config returns base.
func TestComposeInitialPrompt_EmptyConfig(t *testing.T) {
	workerID := "worker-initial-empty"
	config := &WorkflowConfig{}

	rolePrompts := GetPrompts(AgentTypeResearcher)
	expectedBase := rolePrompts.InitialPrompt(workerID)

	result := ComposeInitialPrompt(workerID, AgentTypeResearcher, config)

	require.Equal(t, expectedBase, result,
		"Empty config should return base initial prompt")
}

// ============================================================================
// Edge Cases
// ============================================================================

// TestComposeSystemPrompt_UnknownAgentType verifies fallback to generic.
func TestComposeSystemPrompt_UnknownAgentType(t *testing.T) {
	workerID := "worker-unknown"
	unknownType := AgentType("unknown-type-xyz")

	// Should fall back to generic
	genericPrompts := GetPrompts(AgentTypeGeneric)
	expected := genericPrompts.SystemPrompt(workerID)

	result := ComposeSystemPrompt(workerID, unknownType, nil)

	require.Equal(t, expected, result,
		"Unknown type should fall back to generic prompts")
}

// TestComposeInitialPrompt_UnknownAgentType verifies fallback to generic.
func TestComposeInitialPrompt_UnknownAgentType(t *testing.T) {
	workerID := "worker-unknown-initial"
	unknownType := AgentType("unknown-type-abc")

	// Should fall back to generic
	genericPrompts := GetPrompts(AgentTypeGeneric)
	expected := genericPrompts.InitialPrompt(workerID)

	result := ComposeInitialPrompt(workerID, unknownType, nil)

	require.Equal(t, expected, result,
		"Unknown type should fall back to generic initial prompt")
}

// TestComposeSystemPrompt_WhitespaceOnlyAppend verifies whitespace append is applied.
func TestComposeSystemPrompt_WhitespaceOnlyAppend(t *testing.T) {
	workerID := "worker-whitespace"
	appendText := "   " // Whitespace only

	config := &WorkflowConfig{
		SystemPromptAppend: appendText,
	}

	rolePrompts := GetPrompts(AgentTypeGeneric)
	expectedBase := rolePrompts.SystemPrompt(workerID)

	result := ComposeSystemPrompt(workerID, AgentTypeGeneric, config)

	// Whitespace-only append should still be applied (it's non-empty)
	require.Equal(t, expectedBase+"\n\n"+appendText, result,
		"Whitespace-only append should still be applied")
}

// TestComposeSystemPrompt_ConstraintsIgnored verifies constraints don't affect prompts.
func TestComposeSystemPrompt_ConstraintsIgnored(t *testing.T) {
	workerID := "worker-constraints"

	config := &WorkflowConfig{
		Constraints: []string{"constraint1", "constraint2"},
	}

	rolePrompts := GetPrompts(AgentTypeImplementer)
	expectedBase := rolePrompts.SystemPrompt(workerID)

	result := ComposeSystemPrompt(workerID, AgentTypeImplementer, config)

	require.Equal(t, expectedBase, result,
		"Constraints field should not affect system prompt composition")
}
