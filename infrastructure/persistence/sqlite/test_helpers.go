// infrastructure/persistence/sqlite/test_helpers.go
package sqlite

import (
	"database/sql"
	"testing"

	"osbb-accounting/migrations"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// SetupTestDB creates an in-memory SQLite database with all migrations applied.
// The database is automatically closed when the test completes via t.Cleanup().
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    db := SetupTestDB(t)
//	    repo := sqlite.NewUserRepository(db)
//	    // ... test code
//	}
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "failed to create in-memory database")

	// Run all migrations to create full schema
	err = migrations.RunMigrations(db)
	require.NoError(t, err, "failed to run migrations")

	// Automatically close database when test completes
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close test database: %v", err)
		}
	})

	return db
}

// SetupTestDBWithoutMigrations creates an in-memory SQLite database WITHOUT migrations.
// Use this when you need to test migration logic itself or need a completely empty database.
//
// The database is automatically closed when the test completes via t.Cleanup().
func SetupTestDBWithoutMigrations(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "failed to create in-memory database")

	// Enable foreign keys (migrations would do this, but we're skipping them)
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	require.NoError(t, err, "failed to enable foreign keys")

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close test database: %v", err)
		}
	})

	return db
}

// TruncateTable removes all data from the specified table.
// Useful for cleaning up between test cases while keeping the schema.
func TruncateTable(t *testing.T, db *sql.DB, tableName string) {
	t.Helper()

	_, err := db.Exec("DELETE FROM " + tableName)
	require.NoError(t, err, "failed to truncate table %s", tableName)
}
