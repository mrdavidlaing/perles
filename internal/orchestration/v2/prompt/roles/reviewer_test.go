package roles

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Reviewer Prompt Tests
// ============================================================================

// TestReviewerSystemPrompt_ContainsWorkerID verifies workerID appears in output.
func TestReviewerSystemPrompt_ContainsWorkerID(t *testing.T) {
	workerID := "worker-42"
	prompt := ReviewerSystemPrompt(workerID)
	require.Contains(t, prompt, workerID,
		"ReviewerSystemPrompt should contain the workerID")
}

// TestReviewerSystemPrompt_ReturnsNonEmpty verifies prompt is not empty.
func TestReviewerSystemPrompt_ReturnsNonEmpty(t *testing.T) {
	prompt := ReviewerSystemPrompt("worker-1")
	require.NotEmpty(t, prompt,
		"ReviewerSystemPrompt should return non-empty string")
}

// TestReviewerIdlePrompt_ContainsWorkerID verifies workerID appears in output.
func TestReviewerIdlePrompt_ContainsWorkerID(t *testing.T) {
	workerID := "worker-42"
	prompt := ReviewerIdlePrompt(workerID)
	require.Contains(t, prompt, workerID,
		"ReviewerIdlePrompt should contain the workerID")
}

// TestReviewerIdlePrompt_ReturnsNonEmpty verifies prompt is not empty.
func TestReviewerIdlePrompt_ReturnsNonEmpty(t *testing.T) {
	prompt := ReviewerIdlePrompt("worker-1")
	require.NotEmpty(t, prompt,
		"ReviewerIdlePrompt should return non-empty string")
}

// TestReviewerSystemPromptVersion_IsSemver verifies version follows semver format.
func TestReviewerSystemPromptVersion_IsSemver(t *testing.T) {
	// Semver regex pattern (simplified for major.minor.patch)
	semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	require.True(t, semverPattern.MatchString(ReviewerSystemPromptVersion),
		"ReviewerSystemPromptVersion %q should follow semver format (x.y.z)",
		ReviewerSystemPromptVersion)
}

// TestReviewerSystemPrompt_ContainsReviewCriteria verifies review criteria are documented.
func TestReviewerSystemPrompt_ContainsReviewCriteria(t *testing.T) {
	prompt := ReviewerSystemPrompt("worker-1")

	// Should mention review specialization
	require.Contains(t, prompt, "review",
		"ReviewerSystemPrompt should mention review")

	// Should contain review criteria categories
	criteria := []string{
		"Correctness & Logic",
		"Security",
		"Best Practices",
		"Testing",
	}

	for _, criterion := range criteria {
		require.Contains(t, prompt, criterion,
			"ReviewerSystemPrompt should contain review criterion %q", criterion)
	}
}

// TestReviewerSystemPrompt_ContainsSecurityChecks verifies security review items.
func TestReviewerSystemPrompt_ContainsSecurityChecks(t *testing.T) {
	prompt := ReviewerSystemPrompt("worker-1")

	securityItems := []string{
		"injection",
		"validation",
		"secrets",
	}

	for _, item := range securityItems {
		require.Contains(t, prompt, item,
			"ReviewerSystemPrompt should mention security check %q", item)
	}
}

// TestReviewerSystemPrompt_ContainsVerdictGuidelines verifies APPROVE/DENY criteria.
func TestReviewerSystemPrompt_ContainsVerdictGuidelines(t *testing.T) {
	prompt := ReviewerSystemPrompt("worker-1")

	require.Contains(t, prompt, "DENY if",
		"ReviewerSystemPrompt should explain DENY criteria")
	require.Contains(t, prompt, "APPROVE if",
		"ReviewerSystemPrompt should explain APPROVE criteria")
	require.Contains(t, prompt, "report_review_verdict",
		"ReviewerSystemPrompt should mention report_review_verdict tool")
}

// TestReviewerIdlePrompt_IdentifiesAsReviewer verifies role identification.
func TestReviewerIdlePrompt_IdentifiesAsReviewer(t *testing.T) {
	prompt := ReviewerIdlePrompt("worker-1")
	require.Contains(t, prompt, "reviewer",
		"ReviewerIdlePrompt should identify as reviewer")
}
