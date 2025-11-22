package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"osbb-accounting/domain/entity"
	"osbb-accounting/infrastructure/persistence/sqlite"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOSBBDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	query := `
	CREATE TABLE osbb (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		edrpou TEXT NOT NULL UNIQUE,
		legal_address TEXT NOT NULL,
		actual_address TEXT,
		phone TEXT,
		email TEXT,
		website TEXT,
		chairman_name TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	`
	_, err = db.Exec(query)
	require.NoError(t, err)

	return db
}

func TestOSBBRepository_Create(t *testing.T) {
	db := setupOSBBDB(t)
	defer db.Close()

	repo := sqlite.NewOSBBRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		osbb, err := entity.NewOSBB("Test OSBB", "12345678", "Addr", "Chair", nil, nil, nil, nil)
		require.NoError(t, err)

		err = repo.Create(ctx, osbb)
		require.NoError(t, err)
		assert.Equal(t, int64(1), osbb.ID)
	})

	t.Run("Already Exists", func(t *testing.T) {
		osbb, _ := entity.NewOSBB("Another", "87654321", "Addr", "Chair", nil, nil, nil, nil)
		err := repo.Create(ctx, osbb)
		assert.Error(t, err) // Should fail because ID=1 already exists
	})
}

func TestOSBBRepository_Get(t *testing.T) {
	db := setupOSBBDB(t)
	defer db.Close()

	repo := sqlite.NewOSBBRepository(db)
	ctx := context.Background()

	t.Run("Not Found", func(t *testing.T) {
		_, err := repo.Get(ctx)
		assert.Error(t, err)
	})

	t.Run("Found", func(t *testing.T) {
		osbb, _ := entity.NewOSBB("Test", "12345678", "Addr", "Chair", nil, nil, nil, nil)
		repo.Create(ctx, osbb)

		found, err := repo.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, osbb.Name, found.Name)
	})
}

func TestOSBBRepository_Update(t *testing.T) {
	db := setupOSBBDB(t)
	defer db.Close()

	repo := sqlite.NewOSBBRepository(db)
	ctx := context.Background()

	osbb, _ := entity.NewOSBB("Test", "12345678", "Addr", "Chair", nil, nil, nil, nil)
	repo.Create(ctx, osbb)

	t.Run("Success", func(t *testing.T) {
		osbb.Name = "Updated Name"
		err := repo.Update(ctx, osbb)
		require.NoError(t, err)

		updated, _ := repo.Get(ctx)
		assert.Equal(t, "Updated Name", updated.Name)
	})
}
