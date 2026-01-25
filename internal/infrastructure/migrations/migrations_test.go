package migrations

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/require"

	// Import ncruces driver - this is the same driver Perles uses
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// TestRunMigrations_FreshDB verifies all migrations apply to an empty :memory: database.
func TestRunMigrations_FreshDB(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")
	require.NoError(t, err, "ncruces driver should open :memory: database")
	defer db.Close()

	// Run migrations
	err = RunMigrations(db)
	require.NoError(t, err, "RunMigrations should succeed on fresh database")

	// Verify sessions table was created
	var tableName string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='sessions'`).Scan(&tableName)
	require.NoError(t, err, "sessions table should exist")
	require.Equal(t, "sessions", tableName)
}

// TestRunMigrations_Idempotent verifies calling RunMigrations twice doesn't error.
func TestRunMigrations_Idempotent(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")
	require.NoError(t, err)
	defer db.Close()

	// First run
	err = RunMigrations(db)
	require.NoError(t, err, "first migration run should succeed")

	// Second run should not error (ErrNoChange handled internally)
	err = RunMigrations(db)
	require.NoError(t, err, "second migration run should not error")

	// Verify table still exists
	var tableName string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='sessions'`).Scan(&tableName)
	require.NoError(t, err)
	require.Equal(t, "sessions", tableName)
}

// TestMigrations_Schema verifies sessions table exists with correct columns and indexes.
func TestMigrations_Schema(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")
	require.NoError(t, err)
	defer db.Close()

	err = RunMigrations(db)
	require.NoError(t, err)

	// Verify table has expected columns
	rows, err := db.Query(`PRAGMA table_info(sessions)`)
	require.NoError(t, err)
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt interface{}
		err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk)
		require.NoError(t, err)
		columns[name] = true
	}
	require.NoError(t, rows.Err())

	expectedColumns := []string{"id", "guid", "project", "state", "created_at", "updated_at", "deleted_at"}
	for _, col := range expectedColumns {
		require.True(t, columns[col], "column %s should exist", col)
	}

	// Verify indexes were created
	indexRows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='sessions'`)
	require.NoError(t, err)
	defer indexRows.Close()

	indexes := make(map[string]bool)
	for indexRows.Next() {
		var name string
		require.NoError(t, indexRows.Scan(&name))
		indexes[name] = true
	}
	require.NoError(t, indexRows.Err())

	expectedIndexes := []string{
		"idx_sessions_project",
		"idx_sessions_guid",
		"idx_sessions_deleted_at",
		"idx_sessions_project_state",
	}
	for _, idx := range expectedIndexes {
		require.True(t, indexes[idx], "index %s should exist", idx)
	}
}

// TestMigrations_Down verifies down migration rolls back schema correctly.
func TestMigrations_Down(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")
	require.NoError(t, err)
	defer db.Close()

	// Apply migrations first using the lower-level API for down testing
	driver, err := WithInstance(db, &Config{})
	require.NoError(t, err)

	source, err := iofs.New(MigrationsFS(), ".")
	require.NoError(t, err)

	m, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	require.NoError(t, err)

	err = m.Up()
	require.NoError(t, err, "migrations should apply")

	// Verify table exists before down
	var tableExists bool
	err = db.QueryRow(`SELECT COUNT(*) > 0 FROM sqlite_master WHERE type='table' AND name='sessions'`).Scan(&tableExists)
	require.NoError(t, err)
	require.True(t, tableExists, "sessions table should exist before down migration")

	// Run down migrations
	err = m.Down()
	require.NoError(t, err, "down migrations should succeed")

	// Verify table no longer exists
	err = db.QueryRow(`SELECT COUNT(*) > 0 FROM sqlite_master WHERE type='table' AND name='sessions'`).Scan(&tableExists)
	require.NoError(t, err)
	require.False(t, tableExists, "sessions table should be dropped after down migration")

	// Verify indexes are gone too
	var indexCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND tbl_name='sessions'`).Scan(&indexCount)
	require.NoError(t, err)
	require.Equal(t, 0, indexCount, "all indexes should be dropped")
}

// TestMigrationsFS_Embedded verifies SQL files load from embedded FS at build time.
func TestMigrationsFS_Embedded(t *testing.T) {
	fs := MigrationsFS()
	require.NotNil(t, fs, "MigrationsFS should return non-nil filesystem")

	// Verify we can read directory
	entries, err := embeddedMigrationsFS.ReadDir(".")
	require.NoError(t, err, "should read embedded directory")

	fileNames := make(map[string]bool)
	for _, entry := range entries {
		fileNames[entry.Name()] = true
	}

	require.True(t, fileNames["000001_create_sessions.up.sql"], "up migration should be embedded")
	require.True(t, fileNames["000001_create_sessions.down.sql"], "down migration should be embedded")

	// Read content to verify it's not empty
	upContent, err := embeddedMigrationsFS.ReadFile("000001_create_sessions.up.sql")
	require.NoError(t, err)
	require.Contains(t, string(upContent), "CREATE TABLE sessions")

	downContent, err := embeddedMigrationsFS.ReadFile("000001_create_sessions.down.sql")
	require.NoError(t, err)
	require.Contains(t, string(downContent), "DROP TABLE")
}

// TestNCrucesDriverWithGolangMigrate validates that our custom NCrucesSqlite driver
// works with golang-migrate's migration framework using ncruces/go-sqlite3.
func TestNCrucesDriverWithGolangMigrate(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")
	require.NoError(t, err, "ncruces driver should open :memory: database")
	defer db.Close()

	// Verify connection works
	err = db.Ping()
	require.NoError(t, err, "database should respond to ping")

	// Create our custom ncruces-compatible driver
	driver, err := WithInstance(db, &Config{})
	require.NoError(t, err, "WithInstance should accept ncruces *sql.DB")
	require.NotNil(t, driver, "driver should not be nil")
}

// TestMigrateUp verifies that embedded migrations run successfully using lower-level API.
func TestMigrateUp(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")
	require.NoError(t, err)
	defer db.Close()

	driver, err := WithInstance(db, &Config{})
	require.NoError(t, err)

	source, err := iofs.New(MigrationsFS(), ".")
	require.NoError(t, err, "iofs should load embedded SQL files")

	m, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	require.NoError(t, err, "migrate instance should be created")

	err = m.Up()
	require.NoError(t, err, "migrations should apply successfully")

	// Verify sessions table was created
	var tableName string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='sessions'`).Scan(&tableName)
	require.NoError(t, err, "sessions table should exist")
	require.Equal(t, "sessions", tableName)
}

// TestMigrateIdempotent verifies that running migrations twice handles ErrNoChange.
func TestMigrateIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")
	require.NoError(t, err)
	defer db.Close()

	// First migration run
	driver1, err := WithInstance(db, &Config{})
	require.NoError(t, err)

	source1, err := iofs.New(MigrationsFS(), ".")
	require.NoError(t, err)

	m1, err := migrate.NewWithInstance("iofs", source1, "sqlite3", driver1)
	require.NoError(t, err)

	err = m1.Up()
	require.NoError(t, err, "first migration run should succeed")

	// Close and recreate migrator (simulates app restart)
	driver2, err := WithInstance(db, &Config{})
	require.NoError(t, err)

	source2, err := iofs.New(MigrationsFS(), ".")
	require.NoError(t, err)

	m2, err := migrate.NewWithInstance("iofs", source2, "sqlite3", driver2)
	require.NoError(t, err)

	// Second migration run should return ErrNoChange
	err = m2.Up()
	if err != nil {
		require.True(t, errors.Is(err, migrate.ErrNoChange),
			"second migration run should return ErrNoChange, got: %v", err)
	}
}

// TestInsertAndQueryWithMigration verifies the migrated schema works for CRUD.
func TestInsertAndQueryWithMigration(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")
	require.NoError(t, err)
	defer db.Close()

	err = RunMigrations(db)
	require.NoError(t, err)

	// Insert a test session
	result, err := db.Exec(`
		INSERT INTO sessions (guid, project, state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "test-guid-123", "my-project", "running", 1706000000, 1706000000)
	require.NoError(t, err, "insert should succeed")

	id, err := result.LastInsertId()
	require.NoError(t, err)
	require.Equal(t, int64(1), id, "first insert should have ID 1")

	// Query back
	var guid, project, state string
	var createdAt, updatedAt int64
	var deletedAt *int64
	err = db.QueryRow(`
		SELECT guid, project, state, created_at, updated_at, deleted_at
		FROM sessions WHERE id = ?
	`, id).Scan(&guid, &project, &state, &createdAt, &updatedAt, &deletedAt)
	require.NoError(t, err)
	require.Equal(t, "test-guid-123", guid)
	require.Equal(t, "my-project", project)
	require.Equal(t, "running", state)
	require.Nil(t, deletedAt)

	// Test state CHECK constraint
	_, err = db.Exec(`
		INSERT INTO sessions (guid, project, state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "test-guid-456", "my-project", "invalid_state", 1706000000, 1706000000)
	require.Error(t, err, "CHECK constraint should reject invalid state")
}
