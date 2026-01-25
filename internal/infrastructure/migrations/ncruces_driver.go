// Package migrations provides database migration support for Perles.
//
// This package contains a custom SQLite migration driver compatible with
// ncruces/go-sqlite3 (CGO-free). The standard golang-migrate/v4/database/sqlite3
// driver cannot be used because it imports github.com/mattn/go-sqlite3, which
// causes a driver registration collision (both drivers register as "sqlite3").
//
// The custom driver is based on golang-migrate's sqlite3 driver but removes
// the mattn dependency, allowing use with any sql.DB connection opened via
// the ncruces driver.
package migrations

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/golang-migrate/migrate/v4/database"
)

// DefaultMigrationsTable is the default table name for migration tracking.
const DefaultMigrationsTable = "schema_migrations"

// ErrDatabaseDirty indicates the database has a dirty migration state.
var ErrDatabaseDirty = errors.New("database is dirty")

// ErrNilConfig indicates no config was provided.
var ErrNilConfig = errors.New("no config")

// Config holds configuration for the SQLite migration driver.
type Config struct {
	MigrationsTable string
	NoTxWrap        bool
}

// NCrucesSqlite is a golang-migrate compatible driver for ncruces/go-sqlite3.
// It implements the database.Driver interface without importing mattn/go-sqlite3.
type NCrucesSqlite struct {
	db       *sql.DB
	isLocked atomic.Bool
	config   *Config
}

// WithInstance creates a new NCrucesSqlite driver from an existing sql.DB connection.
// The connection should be opened with ncruces/go-sqlite3 driver.
func WithInstance(instance *sql.DB, config *Config) (database.Driver, error) {
	if config == nil {
		return nil, ErrNilConfig
	}

	if err := instance.Ping(); err != nil {
		return nil, err
	}

	if len(config.MigrationsTable) == 0 {
		config.MigrationsTable = DefaultMigrationsTable
	}

	mx := &NCrucesSqlite{
		db:     instance,
		config: config,
	}
	if err := mx.ensureVersionTable(); err != nil {
		return nil, err
	}
	return mx, nil
}

// ensureVersionTable creates the migrations tracking table if it doesn't exist.
func (m *NCrucesSqlite) ensureVersionTable() (err error) {
	if err = m.Lock(); err != nil {
		return err
	}

	defer func() {
		if e := m.Unlock(); e != nil {
			err = errors.Join(err, e)
		}
	}()

	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (version uint64, dirty bool);
	CREATE UNIQUE INDEX IF NOT EXISTS version_unique ON %s (version);
	`, m.config.MigrationsTable, m.config.MigrationsTable)

	if _, err := m.db.Exec(query); err != nil {
		return err
	}
	return nil
}

// Open is not implemented since we use WithInstance with pre-opened connections.
// This satisfies the database.Driver interface.
func (m *NCrucesSqlite) Open(_ string) (database.Driver, error) {
	return nil, errors.New("Open not implemented; use WithInstance with ncruces driver")
}

// Close closes the database connection.
func (m *NCrucesSqlite) Close() error {
	return m.db.Close()
}

// Lock acquires an in-memory lock for migration operations.
func (m *NCrucesSqlite) Lock() error {
	if !m.isLocked.CompareAndSwap(false, true) {
		return database.ErrLocked
	}
	return nil
}

// Unlock releases the in-memory lock.
func (m *NCrucesSqlite) Unlock() error {
	if !m.isLocked.CompareAndSwap(true, false) {
		return database.ErrNotLocked
	}
	return nil
}

// Run executes a migration.
func (m *NCrucesSqlite) Run(migration io.Reader) error {
	migr, err := io.ReadAll(migration)
	if err != nil {
		return err
	}
	query := string(migr)

	if m.config.NoTxWrap {
		return m.executeQueryNoTx(query)
	}
	return m.executeQuery(query)
}

func (m *NCrucesSqlite) executeQuery(query string) error {
	tx, err := m.db.Begin()
	if err != nil {
		return &database.Error{OrigErr: err, Err: "transaction start failed"}
	}
	if _, err := tx.Exec(query); err != nil {
		if errRollback := tx.Rollback(); errRollback != nil {
			err = errors.Join(err, errRollback)
		}
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	if err := tx.Commit(); err != nil {
		return &database.Error{OrigErr: err, Err: "transaction commit failed"}
	}
	return nil
}

func (m *NCrucesSqlite) executeQueryNoTx(query string) error {
	if _, err := m.db.Exec(query); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	return nil
}

// SetVersion sets the current migration version.
func (m *NCrucesSqlite) SetVersion(version int, dirty bool) error {
	tx, err := m.db.Begin()
	if err != nil {
		return &database.Error{OrigErr: err, Err: "transaction start failed"}
	}

	query := "DELETE FROM " + m.config.MigrationsTable //nolint:gosec // table name is from trusted config, not user input
	if _, err := tx.Exec(query); err != nil {
		if errRollback := tx.Rollback(); errRollback != nil {
			err = errors.Join(err, errRollback)
		}
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}

	// Also re-write the schema version for nil dirty versions to prevent
	// empty schema version for failed down migration on the first migration
	// See: https://github.com/golang-migrate/migrate/issues/330
	if version >= 0 || (version == database.NilVersion && dirty) {
		query := fmt.Sprintf(`INSERT INTO %s (version, dirty) VALUES (?, ?)`, m.config.MigrationsTable) //nolint:gosec // table name is from trusted config, not user input
		if _, err := tx.Exec(query, version, dirty); err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Join(err, errRollback)
			}
			return &database.Error{OrigErr: err, Query: []byte(query)}
		}
	}

	if err := tx.Commit(); err != nil {
		return &database.Error{OrigErr: err, Err: "transaction commit failed"}
	}

	return nil
}

// Version returns the current migration version.
func (m *NCrucesSqlite) Version() (version int, dirty bool, err error) {
	query := "SELECT version, dirty FROM " + m.config.MigrationsTable + " LIMIT 1"
	err = m.db.QueryRow(query).Scan(&version, &dirty)
	if err != nil {
		return database.NilVersion, false, nil
	}
	return version, dirty, nil
}

// Drop drops all tables in the database.
func (m *NCrucesSqlite) Drop() (err error) {
	query := `SELECT name FROM sqlite_master WHERE type = 'table';`
	tables, err := m.db.Query(query)
	if err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	defer func() {
		if errClose := tables.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	tableNames := make([]string, 0)
	for tables.Next() {
		var tableName string
		if err := tables.Scan(&tableName); err != nil {
			return err
		}
		if len(tableName) > 0 {
			tableNames = append(tableNames, tableName)
		}
	}
	if err := tables.Err(); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}

	for _, t := range tableNames {
		query := "DROP TABLE " + t
		err = m.executeQuery(query)
		if err != nil {
			return &database.Error{OrigErr: err, Query: []byte(query)}
		}
	}
	if len(tableNames) > 0 {
		query := "VACUUM"
		_, err = m.db.Exec(query)
		if err != nil {
			return &database.Error{OrigErr: err, Query: []byte(query)}
		}
	}

	return nil
}
