package migrations_test

import (
	"database/sql"
	"testing"

	"osbb-accounting/migrations"

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
	assert.Equal(t, "006", version)
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

func TestRepairMigrations(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	// Run migrations normally first
	err := migrations.RunMigrations(db)
	require.NoError(t, err)

	// Manually insert an orphaned migration record
	_, err = db.Exec("INSERT INTO schema_migrations (version, description, checksum) VALUES (?, ?, ?)", "999", "Orphaned Migration", "fake-checksum")
	require.NoError(t, err)

	// Verify it exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = '999'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Run RepairMigrations
	err = migrations.RepairMigrations(db)
	require.NoError(t, err)

	// Verify it's gone
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = '999'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestIntegrityCheckFailed(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	// Run migrations normally first
	err := migrations.RunMigrations(db)
	require.NoError(t, err)

	// Tamper with a checksum
	_, err = db.Exec("UPDATE schema_migrations SET checksum = 'tampered' WHERE version = '001'")
	require.NoError(t, err)

	// Run migrations again, should fail integrity check
	err = migrations.RunMigrations(db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}
