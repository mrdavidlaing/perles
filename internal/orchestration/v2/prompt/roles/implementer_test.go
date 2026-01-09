package roles

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Implementer Prompt Tests
// ============================================================================

// TestImplementerSystemPrompt_ContainsWorkerID verifies workerID appears in output.
func TestImplementerSystemPrompt_ContainsWorkerID(t *testing.T) {
	workerID := "worker-42"
	prompt := ImplementerSystemPrompt(workerID)
	require.Contains(t, prompt, workerID,
		"ImplementerSystemPrompt should contain the workerID")
}

// TestImplementerSystemPrompt_ReturnsNonEmpty verifies prompt is not empty.
func TestImplementerSystemPrompt_ReturnsNonEmpty(t *testing.T) {
	prompt := ImplementerSystemPrompt("worker-1")
	require.NotEmpty(t, prompt,
		"ImplementerSystemPrompt should return non-empty string")
}

// TestImplementerIdlePrompt_ContainsWorkerID verifies workerID appears in output.
func TestImplementerIdlePrompt_ContainsWorkerID(t *testing.T) {
	workerID := "worker-42"
	prompt := ImplementerIdlePrompt(workerID)
	require.Contains(t, prompt, workerID,
		"ImplementerIdlePrompt should contain the workerID")
}

// TestImplementerIdlePrompt_ReturnsNonEmpty verifies prompt is not empty.
func TestImplementerIdlePrompt_ReturnsNonEmpty(t *testing.T) {
	prompt := ImplementerIdlePrompt("worker-1")
	require.NotEmpty(t, prompt,
		"ImplementerIdlePrompt should return non-empty string")
}

// TestImplementerSystemPromptVersion_IsSemver verifies version follows semver format.
func TestImplementerSystemPromptVersion_IsSemver(t *testing.T) {
	// Semver regex pattern (simplified for major.minor.patch)
	semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	require.True(t, semverPattern.MatchString(ImplementerSystemPromptVersion),
		"ImplementerSystemPromptVersion %q should follow semver format (x.y.z)",
		ImplementerSystemPromptVersion)
}

// TestImplementerSystemPrompt_ContainsTaskGuidelines verifies implementation guidelines.
func TestImplementerSystemPrompt_ContainsTaskGuidelines(t *testing.T) {
	prompt := ImplementerSystemPrompt("worker-1")

	// Should mention implementation specialization
	require.Contains(t, prompt, "implementation",
		"ImplementerSystemPrompt should mention implementation")

	// Should contain implementation guidelines
	guidelines := []string{
		"Understand Before Coding",
		"Write Clean Code",
		"Test Thoroughly",
		"Avoid Anti-Patterns",
	}

	for _, guideline := range guidelines {
		require.Contains(t, prompt, guideline,
			"ImplementerSystemPrompt should contain guideline %q", guideline)
	}
}

// TestImplementerSystemPrompt_ContainsAntiPatterns verifies anti-pattern warnings.
func TestImplementerSystemPrompt_ContainsAntiPatterns(t *testing.T) {
	prompt := ImplementerSystemPrompt("worker-1")

	antiPatterns := []string{
		"test-only helpers",
		"dead code",
		"swallowed errors",
	}

	for _, antiPattern := range antiPatterns {
		require.Contains(t, prompt, antiPattern,
			"ImplementerSystemPrompt should warn about %q", antiPattern)
	}
}

// TestImplementerIdlePrompt_IdentifiesAsImplementer verifies role identification.
func TestImplementerIdlePrompt_IdentifiesAsImplementer(t *testing.T) {
	prompt := ImplementerIdlePrompt("worker-1")
	require.Contains(t, prompt, "implementer",
		"ImplementerIdlePrompt should identify as implementer")
}
