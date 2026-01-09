package roles

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Researcher Prompt Tests
// ============================================================================

// TestResearcherSystemPrompt_ContainsWorkerID verifies workerID appears in output.
func TestResearcherSystemPrompt_ContainsWorkerID(t *testing.T) {
	workerID := "worker-42"
	prompt := ResearcherSystemPrompt(workerID)
	require.Contains(t, prompt, workerID,
		"ResearcherSystemPrompt should contain the workerID")
}

// TestResearcherSystemPrompt_ReturnsNonEmpty verifies prompt is not empty.
func TestResearcherSystemPrompt_ReturnsNonEmpty(t *testing.T) {
	prompt := ResearcherSystemPrompt("worker-1")
	require.NotEmpty(t, prompt,
		"ResearcherSystemPrompt should return non-empty string")
}

// TestResearcherIdlePrompt_ContainsWorkerID verifies workerID appears in output.
func TestResearcherIdlePrompt_ContainsWorkerID(t *testing.T) {
	workerID := "worker-42"
	prompt := ResearcherIdlePrompt(workerID)
	require.Contains(t, prompt, workerID,
		"ResearcherIdlePrompt should contain the workerID")
}

// TestResearcherIdlePrompt_ReturnsNonEmpty verifies prompt is not empty.
func TestResearcherIdlePrompt_ReturnsNonEmpty(t *testing.T) {
	prompt := ResearcherIdlePrompt("worker-1")
	require.NotEmpty(t, prompt,
		"ResearcherIdlePrompt should return non-empty string")
}

// TestResearcherSystemPromptVersion_IsSemver verifies version follows semver format.
func TestResearcherSystemPromptVersion_IsSemver(t *testing.T) {
	// Semver regex pattern (simplified for major.minor.patch)
	semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	require.True(t, semverPattern.MatchString(ResearcherSystemPromptVersion),
		"ResearcherSystemPromptVersion %q should follow semver format (x.y.z)",
		ResearcherSystemPromptVersion)
}

// TestResearcherSystemPrompt_ContainsExplorationGuidelines verifies research guidelines.
func TestResearcherSystemPrompt_ContainsExplorationGuidelines(t *testing.T) {
	prompt := ResearcherSystemPrompt("worker-1")

	// Should mention research specialization
	require.Contains(t, prompt, "Research",
		"ResearcherSystemPrompt should mention Research")

	// Should contain exploration guidelines
	guidelines := []string{
		"Exploration Strategy",
		"Pattern Recognition",
		"Documentation Quality",
		"Analysis Depth",
	}

	for _, guideline := range guidelines {
		require.Contains(t, prompt, guideline,
			"ResearcherSystemPrompt should contain guideline %q", guideline)
	}
}

// TestResearcherSystemPrompt_ContainsOutputFormat verifies output format guidance.
func TestResearcherSystemPrompt_ContainsOutputFormat(t *testing.T) {
	prompt := ResearcherSystemPrompt("worker-1")

	outputSections := []string{
		"Summary",
		"Key Files",
		"Patterns Found",
		"Architecture Notes",
		"Recommendations",
	}

	for _, section := range outputSections {
		require.Contains(t, prompt, section,
			"ResearcherSystemPrompt should describe output section %q", section)
	}
}

// TestResearcherIdlePrompt_IdentifiesAsResearcher verifies role identification.
func TestResearcherIdlePrompt_IdentifiesAsResearcher(t *testing.T) {
	prompt := ResearcherIdlePrompt("worker-1")
	require.Contains(t, prompt, "researcher",
		"ResearcherIdlePrompt should identify as researcher")
}
