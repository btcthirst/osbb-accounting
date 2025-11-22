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

func TestDefaultConfig(t *testing.T) {
	config := sqlite.DefaultConfig()
	assert.NotEmpty(t, config.Path)
	assert.Greater(t, config.MaxOpenConns, 0)
}
