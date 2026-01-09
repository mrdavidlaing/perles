// Package roles provides agent type definitions and prompt registry for specialized AI agents.
package roles

import "errors"

// ErrAgentTypeNotFound is returned when an agent type is not found in the registry.
// Note: GetPrompts falls back to generic, so this is mainly for explicit validation.
var ErrAgentTypeNotFound = errors.New("agent type not found in registry")
