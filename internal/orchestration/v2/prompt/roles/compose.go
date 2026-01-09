// Package roles provides agent type definitions and prompt templates for specialized agents.
// This file contains prompt composition functions that implement three-tier prompt resolution:
// embedded defaults → workflow overrides → runtime composition.
package roles

// WorkflowConfig contains workflow-specific prompt customizations.
// This mirrors workflow.AgentRoleConfig to avoid import cycles, since the workflow
// package imports roles for AgentType validation.
type WorkflowConfig struct {
	// SystemPromptAppend is appended to the default system prompt.
	SystemPromptAppend string

	// SystemPromptOverride completely replaces the default system prompt.
	SystemPromptOverride string

	// Constraints are behavioral constraints (currently unused in prompt composition).
	Constraints []string
}

// ComposeSystemPrompt composes the final system prompt using three-tier resolution:
//  1. Get base prompt from the roles registry for the agent type
//  2. If workflowConfig is nil, return the base prompt
//  3. If workflowConfig.SystemPromptOverride is set, return the override
//  4. If workflowConfig.SystemPromptAppend is set, return base + append
//  5. Otherwise, return the base prompt
func ComposeSystemPrompt(workerID string, agentType AgentType, workflowConfig *WorkflowConfig) string {
	// Get base prompt from registry (falls back to generic if type not found)
	rolePrompts := GetPrompts(agentType)
	basePrompt := rolePrompts.SystemPrompt(workerID)

	if workflowConfig == nil {
		return basePrompt
	}

	if workflowConfig.SystemPromptOverride != "" {
		return workflowConfig.SystemPromptOverride
	}

	if workflowConfig.SystemPromptAppend != "" {
		return basePrompt + "\n\n" + workflowConfig.SystemPromptAppend
	}

	return basePrompt
}

// ComposeInitialPrompt composes the final initial/idle prompt using three-tier resolution:
//  1. Get base prompt from the roles registry for the agent type
//  2. If workflowConfig is nil, return the base prompt
//  3. If workflowConfig has no customizations, return the base prompt
//
// Note: Initial prompts currently only support the base tier. Workflow customization
// for initial prompts may be added in the future if needed.
func ComposeInitialPrompt(workerID string, agentType AgentType, workflowConfig *WorkflowConfig) string {
	rolePrompts := GetPrompts(agentType)
	basePrompt := rolePrompts.InitialPrompt(workerID)

	// TODO Initial prompts don't currently support workflow customization
	// The workflow config is accepted for API consistency and future extensibility
	_ = workflowConfig

	return basePrompt
}
