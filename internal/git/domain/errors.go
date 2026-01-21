package domain

import "errors"

// Git-specific errors for worktree operations.
var (
	// ErrBranchAlreadyCheckedOut indicates the branch is checked out in another worktree.
	ErrBranchAlreadyCheckedOut = errors.New("branch already checked out in another worktree")

	// ErrPathAlreadyExists indicates the worktree path already exists.
	ErrPathAlreadyExists = errors.New("worktree path already exists")

	// ErrWorktreeLocked indicates the worktree is locked.
	ErrWorktreeLocked = errors.New("worktree is locked")

	// ErrNotGitRepo indicates the directory is not a git repository.
	ErrNotGitRepo = errors.New("not a git repository")

	// ErrUnsafeParentDirectory indicates the parent directory is restricted.
	ErrUnsafeParentDirectory = errors.New("unsafe parent directory")

	// ErrDetachedHead indicates HEAD is not pointing to a branch (detached HEAD state).
	ErrDetachedHead = errors.New("detached HEAD state")

	// ErrInvalidBranchName indicates the branch name format is invalid per git check-ref-format.
	ErrInvalidBranchName = errors.New("invalid branch name format")

	// ErrWorktreeTimeout is returned when a git worktree operation times out.
	ErrWorktreeTimeout = errors.New("git worktree timed out")

	// ErrDiffTimeout is returned when a git diff operation times out.
	ErrDiffTimeout = errors.New("git diff timed out")
)
