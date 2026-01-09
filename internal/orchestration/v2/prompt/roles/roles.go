package roles

import (
	"strings"
)

// AgentType represents the specialization of an AI agent.
// This is a string-based enum for type safety and serialization.
type AgentType string

const (
	// AgentTypeGeneric is the default agent type with no specialization.
	// Used when no specific agent type is requested.
	AgentTypeGeneric AgentType = ""

	// AgentTypeImplementer specializes in code implementation and testing.
	AgentTypeImplementer AgentType = "implementer"

	// AgentTypeReviewer specializes in code review and quality assessment.
	AgentTypeReviewer AgentType = "reviewer"

	// AgentTypeResearcher specializes in codebase exploration and analysis.
	AgentTypeResearcher AgentType = "researcher"
)

// knownAgentTypes is the set of valid agent types for validation.
var knownAgentTypes = map[AgentType]bool{
	AgentTypeGeneric:     true,
	AgentTypeImplementer: true,
	AgentTypeReviewer:    true,
	AgentTypeResearcher:  true,
}

// IsValid returns true if the agent type is a known valid type.
// It also performs security validation to reject shell injection and path traversal attempts.
func (t AgentType) IsValid() bool {
	// Security: reject strings with shell metacharacters
	s := string(t)
	if strings.ContainsAny(s, ";|&$`\\\"'<>(){}[]!#*?~") {
		return false
	}

	// Security: reject path traversal attempts
	if strings.Contains(s, "..") || strings.Contains(s, "/") {
		return false
	}

	// Check against known types
	return knownAgentTypes[t]
}

// String returns the string representation of the agent type.
func (t AgentType) String() string {
	if t == AgentTypeGeneric {
		return "generic"
	}
	return string(t)
}

// RolePrompts contains the prompt templates for a specific agent type.
type RolePrompts struct {
	// SystemPrompt returns the system prompt for the agent.
	// The workerID parameter identifies the worker instance.
	SystemPrompt func(workerID string) string

	// InitialPrompt returns the initial user prompt for the agent.
	// The workerID parameter identifies the worker instance.
	InitialPrompt func(workerID string) string
}

// Registry maps agent types to their prompt templates.
// This is populated by the individual prompt files (generic.go, implementer.go, etc.)
var Registry = make(map[AgentType]RolePrompts)

// GetPrompts returns the prompts for the given agent type.
// If the agent type is not found, it falls back to generic prompts.
// This ensures graceful degradation for unknown types.
func GetPrompts(agentType AgentType) RolePrompts {
	if prompts, ok := Registry[agentType]; ok {
		return prompts
	}
	// Fall back to generic for unknown types
	return Registry[AgentTypeGeneric]
}
