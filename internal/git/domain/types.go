// Package git provides domain types for git operations.
package domain

import "time"

// BranchInfo holds information about a git branch.
type BranchInfo struct {
	Name      string // Branch name (e.g., "main", "feature/auth")
	IsCurrent bool   // True if this is the currently checked out branch
}

// CommitInfo holds information about a git commit.
type CommitInfo struct {
	Hash      string    // Full 40-char SHA
	ShortHash string    // 7-char abbreviated hash
	Subject   string    // First line of commit message
	Author    string    // Author name
	Date      time.Time // Commit timestamp
	IsPushed  bool      // True if commit exists on the remote tracking branch
}

// WorktreeInfo holds information about a git worktree.
type WorktreeInfo struct {
	Path   string
	Branch string
	HEAD   string
}
