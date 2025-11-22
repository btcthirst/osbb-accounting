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

func setupRoleDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	query := `
	CREATE TABLE roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		description TEXT,
		is_system INTEGER NOT NULL DEFAULT 0,
		is_active INTEGER NOT NULL DEFAULT 1,
		deleted_at INTEGER,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	`
	_, err = db.Exec(query)
	require.NoError(t, err)

	return db
}

func TestRoleRepository_Create(t *testing.T) {
	db := setupRoleDB(t)
	defer db.Close()

	repo := sqlite.NewRoleRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		role, err := entity.NewRole("Admin", nil, false)
		require.NoError(t, err)

		err = repo.Create(ctx, role)
		require.NoError(t, err)
		assert.NotZero(t, role.ID)
	})

	t.Run("Duplicate Name", func(t *testing.T) {
		r1, _ := entity.NewRole("User", nil, false)
		repo.Create(ctx, r1)

		r2, _ := entity.NewRole("User", nil, false)
		err := repo.Create(ctx, r2)
		assert.Error(t, err)
	})
}

func TestRoleRepository_GetByName(t *testing.T) {
	db := setupRoleDB(t)
	defer db.Close()

	repo := sqlite.NewRoleRepository(db)
	ctx := context.Background()

	role, _ := entity.NewRole("Manager", nil, false)
	repo.Create(ctx, role)

	found, err := repo.GetByName(ctx, "Manager")
	require.NoError(t, err)
	assert.Equal(t, role.ID, found.ID)
}
