package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionNotFoundError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *SessionNotFoundError
		expected string
	}{
		{
			name:     "basic case",
			err:      &SessionNotFoundError{GUID: "abc-123", Project: "my-project"},
			expected: `session not found: guid="abc-123" project="my-project"`,
		},
		{
			name:     "empty values",
			err:      &SessionNotFoundError{GUID: "", Project: ""},
			expected: `session not found: guid="" project=""`,
		},
		{
			name:     "special characters",
			err:      &SessionNotFoundError{GUID: "guid/with/slashes", Project: "project-with-dashes"},
			expected: `session not found: guid="guid/with/slashes" project="project-with-dashes"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestSessionNotFoundError_ImplementsError(t *testing.T) {
	var err error = &SessionNotFoundError{GUID: "guid", Project: "project"}
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "session not found")
}

func TestNoActiveSessionError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *NoActiveSessionError
		expected string
	}{
		{
			name:     "basic case",
			err:      &NoActiveSessionError{Project: "my-project"},
			expected: `no active session for project "my-project"`,
		},
		{
			name:     "empty project",
			err:      &NoActiveSessionError{Project: ""},
			expected: `no active session for project ""`,
		},
		{
			name:     "project with special chars",
			err:      &NoActiveSessionError{Project: "org/repo"},
			expected: `no active session for project "org/repo"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestNoActiveSessionError_ImplementsError(t *testing.T) {
	var err error = &NoActiveSessionError{Project: "project"}
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "no active session")
}

func TestActiveSessionExistsError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ActiveSessionExistsError
		expected string
	}{
		{
			name:     "basic case",
			err:      &ActiveSessionExistsError{GUID: "abc-123", Project: "my-project"},
			expected: `active session already exists: guid="abc-123" project="my-project"`,
		},
		{
			name:     "empty values",
			err:      &ActiveSessionExistsError{GUID: "", Project: ""},
			expected: `active session already exists: guid="" project=""`,
		},
		{
			name:     "uuid format",
			err:      &ActiveSessionExistsError{GUID: "550e8400-e29b-41d4-a716-446655440000", Project: "perles"},
			expected: `active session already exists: guid="550e8400-e29b-41d4-a716-446655440000" project="perles"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestActiveSessionExistsError_ImplementsError(t *testing.T) {
	var err error = &ActiveSessionExistsError{GUID: "guid", Project: "project"}
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "active session already exists")
}

func TestErrorTypes_DistinctMessages(t *testing.T) {
	notFound := &SessionNotFoundError{GUID: "guid", Project: "project"}
	noActive := &NoActiveSessionError{Project: "project"}
	alreadyExists := &ActiveSessionExistsError{GUID: "guid", Project: "project"}

	// Each error type should have a distinct message prefix
	require.Contains(t, notFound.Error(), "not found")
	require.Contains(t, noActive.Error(), "no active")
	require.Contains(t, alreadyExists.Error(), "already exists")

	// Messages should be different from each other
	require.NotEqual(t, notFound.Error(), noActive.Error())
	require.NotEqual(t, notFound.Error(), alreadyExists.Error())
	require.NotEqual(t, noActive.Error(), alreadyExists.Error())
}
