package domain

import "fmt"

// SessionNotFoundError indicates that a session with the specified identifiers
// could not be found in the repository.
type SessionNotFoundError struct {
	GUID    string
	Project string
}

// Error implements the error interface.
func (e *SessionNotFoundError) Error() string {
	return fmt.Sprintf("session not found: guid=%q project=%q", e.GUID, e.Project)
}

// NoActiveSessionError indicates that no session is currently in the running
// state for the specified project.
type NoActiveSessionError struct {
	Project string
}

// Error implements the error interface.
func (e *NoActiveSessionError) Error() string {
	return fmt.Sprintf("no active session for project %q", e.Project)
}

// ActiveSessionExistsError indicates that an attempt was made to create a new
// running session when one already exists for the project.
type ActiveSessionExistsError struct {
	GUID    string
	Project string
}

// Error implements the error interface.
func (e *ActiveSessionExistsError) Error() string {
	return fmt.Sprintf("active session already exists: guid=%q project=%q", e.GUID, e.Project)
}
