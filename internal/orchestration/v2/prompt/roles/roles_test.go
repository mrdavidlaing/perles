package roles

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// AgentType.IsValid() Tests
// ============================================================================

// TestAgentType_IsValid validates known types return true.
func TestAgentType_IsValid(t *testing.T) {
	validTypes := []AgentType{
		AgentTypeGeneric,
		AgentTypeImplementer,
		AgentTypeReviewer,
		AgentTypeResearcher,
	}

	for _, agentType := range validTypes {
		t.Run(agentType.String(), func(t *testing.T) {
			require.True(t, agentType.IsValid(),
				"Expected %q to be valid", agentType)
		})
	}
}

// TestAgentType_IsValid_RejectsUnknown verifies unknown types return false.
func TestAgentType_IsValid_RejectsUnknown(t *testing.T) {
	unknownTypes := []AgentType{
		"unknown",
		"hacker",
		"admin",
		"root",
		"super-agent",
	}

	for _, agentType := range unknownTypes {
		t.Run(string(agentType), func(t *testing.T) {
			require.False(t, agentType.IsValid(),
				"Expected %q to be invalid", agentType)
		})
	}
}

// TestAgentType_IsValid_EmptyString verifies empty string (generic) returns true.
func TestAgentType_IsValid_EmptyString(t *testing.T) {
	emptyType := AgentType("")
	require.True(t, emptyType.IsValid(),
		"Empty string should be valid (maps to AgentTypeGeneric)")
	require.Equal(t, AgentTypeGeneric, emptyType,
		"Empty string should equal AgentTypeGeneric")
}

// ============================================================================
// Security Tests
// ============================================================================

// TestAgentType_RejectsShellInjection verifies strings with shell chars are rejected.
func TestAgentType_RejectsShellInjection(t *testing.T) {
	shellInjectionAttempts := []AgentType{
		"; rm -rf /",
		"| cat /etc/passwd",
		"& whoami",
		"$(malicious)",
		"`malicious`",
		"\\n whoami",
		"\"breakout\"",
		"'injection'",
		"< /etc/passwd",
		"> /tmp/pwned",
		"(subshell)",
		"{expansion}",
		"[brackets]",
		"!history",
		"#comment",
		"*glob",
		"?wildcard",
		"~home",
	}

	for _, attempt := range shellInjectionAttempts {
		t.Run(string(attempt), func(t *testing.T) {
			require.False(t, attempt.IsValid(),
				"Shell injection attempt %q should be rejected", attempt)
		})
	}
}

// TestAgentType_RejectsPathTraversal verifies strings with ../ are rejected.
func TestAgentType_RejectsPathTraversal(t *testing.T) {
	pathTraversalAttempts := []AgentType{
		"../../../etc/passwd",
		"..\\..\\windows",
		"/etc/passwd",
		"path/to/file",
		"implementer/../admin",
		"..",
	}

	for _, attempt := range pathTraversalAttempts {
		t.Run(string(attempt), func(t *testing.T) {
			require.False(t, attempt.IsValid(),
				"Path traversal attempt %q should be rejected", attempt)
		})
	}
}

// ============================================================================
// Registry Tests
// ============================================================================

// TestRegistry_AllTypesPresent verifies all AgentType constants have registry entries.
func TestRegistry_AllTypesPresent(t *testing.T) {
	requiredTypes := []AgentType{
		AgentTypeGeneric,
		AgentTypeImplementer,
		AgentTypeReviewer,
		AgentTypeResearcher,
	}

	for _, agentType := range requiredTypes {
		t.Run(agentType.String(), func(t *testing.T) {
			_, exists := Registry[agentType]
			require.True(t, exists,
				"Registry should contain entry for %q", agentType)
		})
	}
}

// TestGetPrompts_ReturnsCorrectPrompts verifies each type returns correct prompts.
func TestGetPrompts_ReturnsCorrectPrompts(t *testing.T) {
	testCases := []AgentType{
		AgentTypeGeneric,
		AgentTypeImplementer,
		AgentTypeReviewer,
		AgentTypeResearcher,
	}

	for _, agentType := range testCases {
		t.Run(agentType.String(), func(t *testing.T) {
			prompts := GetPrompts(agentType)
			require.NotNil(t, prompts.SystemPrompt,
				"SystemPrompt should not be nil for %q", agentType)
			require.NotNil(t, prompts.InitialPrompt,
				"InitialPrompt should not be nil for %q", agentType)

			// Verify the prompts contain the worker ID and are non-empty
			workerID := "worker-test-123"
			systemPrompt := prompts.SystemPrompt(workerID)
			initialPrompt := prompts.InitialPrompt(workerID)

			require.NotEmpty(t, systemPrompt,
				"SystemPrompt should not be empty for %q", agentType)
			require.NotEmpty(t, initialPrompt,
				"InitialPrompt should not be empty for %q", agentType)
			require.Contains(t, systemPrompt, workerID,
				"SystemPrompt should contain workerID for %q", agentType)
			require.Contains(t, initialPrompt, workerID,
				"InitialPrompt should contain workerID for %q", agentType)
		})
	}
}

// TestGetPrompts_FallbackToGeneric verifies unknown type falls back to generic.
func TestGetPrompts_FallbackToGeneric(t *testing.T) {
	unknownType := AgentType("unknown-type")
	prompts := GetPrompts(unknownType)

	// Should get generic prompts
	genericPrompts := GetPrompts(AgentTypeGeneric)

	// The system prompts should be the same function
	require.Equal(t,
		prompts.SystemPrompt("worker-1"),
		genericPrompts.SystemPrompt("worker-1"),
		"Unknown type should fall back to generic prompts")
}

// ============================================================================
// String() Tests
// ============================================================================

// TestAgentType_String verifies String() returns expected values.
func TestAgentType_String(t *testing.T) {
	testCases := []struct {
		agentType AgentType
		expected  string
	}{
		{AgentTypeGeneric, "generic"},
		{AgentTypeImplementer, "implementer"},
		{AgentTypeReviewer, "reviewer"},
		{AgentTypeResearcher, "researcher"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.agentType.String(),
				"String() should return %q", tc.expected)
		})
	}
}

// ============================================================================
// Error Tests
// ============================================================================

// TestErrAgentTypeNotFound verifies the error is defined and has expected message.
func TestErrAgentTypeNotFound(t *testing.T) {
	require.NotNil(t, ErrAgentTypeNotFound, "ErrAgentTypeNotFound should be defined")
	require.Contains(t, ErrAgentTypeNotFound.Error(), "not found",
		"Error message should mention 'not found'")
}
