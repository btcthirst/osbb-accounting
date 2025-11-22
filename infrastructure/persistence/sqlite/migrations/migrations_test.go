package migrations_test

import (
	"database/sql"
	"testing"

	"osbb-accounting/infrastructure/persistence/sqlite/migrations"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	return db
}

func TestMigrationRunner_New(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	runner, err := migrations.NewMigrationRunner(db)
	require.NoError(t, err)
	assert.NotNil(t, runner)
}

func TestMigrationRunner_Status(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	runner, err := migrations.NewMigrationRunner(db)
	require.NoError(t, err)

	// Ensure migrations table exists for Status to work
	err = migrations.RunMigrations(db)
	require.NoError(t, err)

	err = runner.Status()
	assert.NoError(t, err)
}

func TestMigrationRunner_GetCurrentVersion(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	runner, err := migrations.NewMigrationRunner(db)
	require.NoError(t, err)

	// Run migrations to set version
	err = migrations.RunMigrations(db)
	require.NoError(t, err)

	version, err := runner.GetCurrentVersion()
	require.NoError(t, err)
	assert.Equal(t, "002", version)
}

func TestRunMigrations(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	err := migrations.RunMigrations(db)
	require.NoError(t, err)

	// Check if tables exist
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "users", tableName)
}

func TestResetDatabase(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	// Run migrations first
	err := migrations.RunMigrations(db)
	require.NoError(t, err)

	// Insert some data
	// Password hash must be 60 chars
	hash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	_, err = db.Exec("INSERT INTO users (username, email, first_name, last_name, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)", "test", "test@example.com", "Test", "User", hash, 123, 123)
	require.NoError(t, err)

	// Reset
	err = migrations.ResetDatabase(db)
	require.NoError(t, err)

	// Verify data is gone
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
