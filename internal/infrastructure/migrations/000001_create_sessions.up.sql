-- Sessions table for orchestration session tracking
-- Multi-tenant via project column, single database at ~/.perles/perles.db
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    guid TEXT NOT NULL UNIQUE,
    project TEXT NOT NULL,
    name TEXT,
    state TEXT NOT NULL CHECK(state IN ('pending', 'running', 'paused', 'completed', 'failed', 'timed_out')),
    template_id TEXT,
    epic_id TEXT,
    work_dir TEXT,
    worktree_path TEXT,
    worktree_branch TEXT,
    owner_created_pid INTEGER,
    owner_current_pid INTEGER,
    created_at INTEGER NOT NULL,
    started_at INTEGER,
    paused_at INTEGER,
    updated_at INTEGER NOT NULL,
    archived_at INTEGER,
    deleted_at INTEGER
);

-- Indexes for common query patterns
CREATE INDEX idx_sessions_project ON sessions(project);
CREATE INDEX idx_sessions_guid ON sessions(guid);
CREATE INDEX idx_sessions_deleted_at ON sessions(deleted_at);
CREATE INDEX idx_sessions_archived_at ON sessions(archived_at);
CREATE INDEX idx_sessions_project_state ON sessions(project, state) WHERE deleted_at IS NULL;
