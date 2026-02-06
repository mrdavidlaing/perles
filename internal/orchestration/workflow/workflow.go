// Package workflow provides workflow template management for orchestration mode.
// It supports loading and managing both built-in and user-defined workflow templates.
package workflow

// AgentRoleConfig defines per-agent-type customizations for a workflow.
// This allows workflows to customize prompts and constraints for specific agent types
// (e.g., implementer, reviewer, researcher).
type AgentRoleConfig struct {
	// SystemPromptAppend is appended to the default system prompt for this agent type.
	// Use this to add workflow-specific instructions without replacing the base prompt.
	SystemPromptAppend string

	// SystemPromptOverride completely replaces the default system prompt.
	SystemPromptOverride string

	// Constraints are behavioral constraints for this agent type within the workflow.
	// These are advisory hints that may be included in prompt composition.
	Constraints []string
}

// Source indicates where a workflow template originated from.
type Source int

const (
	// SourceBuiltIn indicates a workflow bundled with the application.
	SourceBuiltIn Source = iota
	// SourceCommunity indicates a community-contributed workflow.
	SourceCommunity
	// SourceUser indicates a workflow from the user's configuration directory.
	SourceUser
)

// String returns a human-readable representation of the Source.
func (s Source) String() string {
	switch s {
	case SourceBuiltIn:
		return "built-in"
	case SourceCommunity:
		return "community"
	case SourceUser:
		return "user"
	default:
		return "unknown"
	}
}

// TargetMode indicates which application mode(s) a workflow template is designed for.
type TargetMode string

const (
	// TargetOrchestration indicates a workflow designed for multi-agent orchestration mode.
	TargetOrchestration TargetMode = "orchestration"
	// TargetChat indicates a workflow designed for single-agent chat mode.
	TargetChat TargetMode = "chat"
	// TargetBoth indicates a workflow usable in both modes (empty string for backwards compatibility).
	TargetBoth TargetMode = ""
)

// Workflow represents a workflow template that can be used in orchestration mode.
type Workflow struct {
	// ID is derived from the filename (e.g., "debate" from "debate.md").
	ID string

	// Name is the human-readable display name from frontmatter.
	Name string

	// Description is a brief description from frontmatter.
	Description string

	// Category is an optional grouping category from frontmatter.
	Category string

	// Workers is the number of workers required by this workflow.
	// A value of 0 (or omitted in frontmatter) indicates lazy spawn mode,
	// where workers are spawned on-demand as needed by the workflow.
	Workers int

	// TargetMode indicates which application mode(s) this workflow is designed for.
	// Empty string (TargetBoth) means the workflow works in both orchestration and chat modes.
	TargetMode TargetMode `yaml:"target_mode"`

	// AgentRoles contains per-agent-type customizations for this workflow.
	// Keys are agent type strings (e.g., "implementer", "reviewer", "researcher").
	// If nil or empty, the workflow uses default prompts for all agent types.
	AgentRoles map[string]AgentRoleConfig

	// Content is the full markdown content (including frontmatter).
	Content string

	// Source indicates whether this is a built-in or user-defined workflow.
	Source Source

	// FilePath is the absolute path for user workflows (empty for built-in).
	FilePath string
}
