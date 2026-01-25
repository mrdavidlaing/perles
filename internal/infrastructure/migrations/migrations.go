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
//
// Usage:
//
//	db, _ := sql.Open("sqlite3", "file:path/to/db.sqlite")
//	err := migrations.RunMigrations(db)
package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var embeddedMigrationsFS embed.FS

// MigrationsFS returns the embedded filesystem containing migration SQL files.
// This can be used for testing or custom migration scenarios.
func MigrationsFS() fs.FS {
	return embeddedMigrationsFS
}

// RunMigrations applies all pending database migrations to the provided database.
// It uses golang-migrate with the embedded SQL files and the custom NCrucesSqlite driver.
//
// The function handles migrate.ErrNoChange gracefully - if all migrations have
// already been applied, it returns nil rather than an error.
//
// Example:
//
//	db, err := sql.Open("sqlite3", "file:~/.perles/perles.db")
//	if err != nil {
//	    return err
//	}
//	if err := migrations.RunMigrations(db); err != nil {
//	    return fmt.Errorf("failed to run migrations: %w", err)
//	}
func RunMigrations(db *sql.DB) error {
	// Create iofs source from embedded filesystem
	source, err := iofs.New(embeddedMigrationsFS, ".")
	if err != nil {
		return err
	}

	// Create our ncruces-compatible driver instance
	driver, err := WithInstance(db, &Config{})
	if err != nil {
		return err
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	if err != nil {
		return err
	}

	// Run migrations
	if err := m.Up(); err != nil {
		// ErrNoChange means all migrations already applied - not an error
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return err
	}

	return nil
}
