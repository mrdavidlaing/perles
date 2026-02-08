-- Add worktree_mode column for three-option worktree isolation (none/new/existing)
ALTER TABLE sessions ADD COLUMN worktree_mode TEXT DEFAULT '';
