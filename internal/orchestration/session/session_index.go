package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// SessionIndexVersion is the current schema version for the session index.
	SessionIndexVersion = "1.0"
)

// SessionIndex tracks all sessions in a perles sessions directory.
type SessionIndex struct {
	// Version is the schema version for forward compatibility.
	Version string `json:"version"`

	// Sessions is the list of all sessions in chronological order.
	Sessions []SessionIndexEntry `json:"sessions"`
}

// SessionIndexEntry contains summary information about a single session.
type SessionIndexEntry struct {
	// ID is the unique session identifier (UUID).
	ID string `json:"id"`

	// StartTime is when the session was created.
	StartTime time.Time `json:"start_time"`

	// EndTime is when the session ended (zero if still running).
	EndTime time.Time `json:"end_time,omitzero"`

	// Status is the session's final status.
	Status Status `json:"status"`

	// EpicID is the bd epic ID associated with this session (if any).
	EpicID string `json:"epic_id,omitempty"`

	// WorkDir is the working directory where the session was started.
	WorkDir string `json:"work_dir"`

	// AccountabilitySummaryPath is the path to the aggregated accountability summary.
	AccountabilitySummaryPath string `json:"accountability_summary_path,omitempty"`

	// WorkerCount is the number of workers that participated in this session.
	WorkerCount int `json:"worker_count"`

	// TasksCompleted is the number of tasks completed during this session.
	TasksCompleted int `json:"tasks_completed"`

	// TotalCommits is the number of commits made during this session.
	TotalCommits int `json:"total_commits"`
}

// LoadSessionIndex loads an existing session index from the given path.
// If the file doesn't exist, it returns an empty index with the current version.
// If the file exists but contains invalid JSON, it returns an error.
func LoadSessionIndex(path string) (*SessionIndex, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: path is trusted input from caller
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty index for missing file
			return &SessionIndex{
				Version:  SessionIndexVersion,
				Sessions: []SessionIndexEntry{},
			}, nil
		}
		return nil, fmt.Errorf("reading session index: %w", err)
	}

	var index SessionIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("parsing session index: %w", err)
	}

	return &index, nil
}

// SaveSessionIndex writes the session index to the given path using atomic rename.
// It writes to a temporary file first, then renames to the final path to ensure
// the file is never in a partially-written state.
func SaveSessionIndex(path string, index *SessionIndex) error {
	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session index: %w", err)
	}

	// Write to temporary file in the same directory (required for atomic rename)
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, "sessions.*.tmp")
	if err != nil {
		return fmt.Errorf("creating temporary session index file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Write data and close the file
	_, writeErr := tmpFile.Write(data)
	closeErr := tmpFile.Close()
	if writeErr != nil {
		_ = os.Remove(tmpPath) // best effort cleanup
		return fmt.Errorf("writing temporary session index: %w", writeErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath) // best effort cleanup
		return fmt.Errorf("closing temporary session index: %w", closeErr)
	}

	// Atomic rename to final path
	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on rename failure
		_ = os.Remove(tmpPath) // best effort cleanup
		return fmt.Errorf("renaming session index: %w", err)
	}

	return nil
}
