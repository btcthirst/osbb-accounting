package sqlite_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"osbb-accounting/infrastructure/persistence/sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "osbb-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	config := &sqlite.Config{
		Path:            dbPath,
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Minute,
	}

	// Test Connect
	db, err := sqlite.Connect(config)
	require.NoError(t, err)
	defer db.Close()

	assert.NotNil(t, db)

	// Verify file exists
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)

	// Test GetVersion
	version, err := sqlite.GetVersion(db)
	require.NoError(t, err)
	assert.NotEmpty(t, version)

	// Test CheckIntegrity
	err = sqlite.CheckIntegrity(db)
	assert.NoError(t, err)

	// Test Vacuum
	err = sqlite.Vacuum(db)
	assert.NoError(t, err)

	// Test Analyze
	err = sqlite.Analyze(db)
	assert.NoError(t, err)

	// Test GetStats
	stats := sqlite.GetStats(db)
	assert.Equal(t, 1, stats.MaxOpenConnections)
}

func TestConnect_Errors(t *testing.T) {
	// Test with invalid path (directory that cannot be created)
	// We use a path under a non-existent directory with no permissions,
	// but since we run as user, we can just use a path that is clearly invalid for mkdir
	// or use a read-only directory if possible.
	// Easiest is to use a file as a directory.

	tempDir, err := os.MkdirTemp("", "osbb-test-error")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a file
	dummyFile := filepath.Join(tempDir, "dummy")
	err = os.WriteFile(dummyFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Try to create a DB inside that file (which should fail mkdir)
	config := &sqlite.Config{
		Path: filepath.Join(dummyFile, "osbb.db"),
	}

	db, err := sqlite.Connect(config)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to create database directory")
}
