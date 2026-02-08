-- SQLite does not support DROP COLUMN, so we recreate the table without worktree_mode.
-- This migration is destructive and loses the worktree_mode data.
CREATE TABLE sessions_backup AS SELECT
    id, guid, project, name, state, template_id, epic_id, work_dir, labels,
    worktree_enabled, worktree_base_branch, worktree_branch_name, worktree_path, worktree_branch, session_dir,
    owner_created_pid, owner_current_pid, tokens_used, active_workers, last_heartbeat_at, last_progress_at,
    created_at, started_at, paused_at, completed_at, updated_at, archived_at, deleted_at
FROM sessions;

DROP TABLE sessions;

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    guid TEXT NOT NULL UNIQUE,
    project TEXT NOT NULL,
    name TEXT,
    state TEXT NOT NULL CHECK(state IN ('pending', 'running', 'paused', 'completed', 'failed', 'timed_out')),
    template_id TEXT,
    epic_id TEXT,
    work_dir TEXT,
    labels TEXT,
    worktree_enabled INTEGER NOT NULL DEFAULT 0,
    worktree_base_branch TEXT,
    worktree_branch_name TEXT,
    worktree_path TEXT,
    worktree_branch TEXT,
    session_dir TEXT,
    owner_created_pid INTEGER,
    owner_current_pid INTEGER,
    tokens_used INTEGER NOT NULL DEFAULT 0,
    active_workers INTEGER NOT NULL DEFAULT 0,
    last_heartbeat_at INTEGER,
    last_progress_at INTEGER,
    created_at INTEGER NOT NULL,
    started_at INTEGER,
    paused_at INTEGER,
    completed_at INTEGER,
    updated_at INTEGER NOT NULL,
    archived_at INTEGER,
    deleted_at INTEGER
);

INSERT INTO sessions SELECT * FROM sessions_backup;
DROP TABLE sessions_backup;

CREATE INDEX idx_sessions_project ON sessions(project);
CREATE INDEX idx_sessions_guid ON sessions(guid);
CREATE INDEX idx_sessions_deleted_at ON sessions(deleted_at);
CREATE INDEX idx_sessions_archived_at ON sessions(archived_at);
CREATE INDEX idx_sessions_project_state ON sessions(project, state) WHERE deleted_at IS NULL;
