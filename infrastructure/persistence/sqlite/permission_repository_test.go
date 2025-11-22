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

func setupPermissionDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	query := `
	CREATE TABLE permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		resource TEXT NOT NULL,
		action TEXT NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at INTEGER NOT NULL
	);
	`
	_, err = db.Exec(query)
	require.NoError(t, err)

	return db
}

func TestPermissionRepository_Create(t *testing.T) {
	db := setupPermissionDB(t)
	defer db.Close()

	repo := sqlite.NewPermissionRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		perm, err := entity.NewPermission("user.create", "Create User", nil, "user", entity.ActionCreate)
		require.NoError(t, err)

		err = repo.Create(ctx, perm)
		require.NoError(t, err)
		assert.NotZero(t, perm.ID)
	})

	t.Run("Duplicate Code", func(t *testing.T) {
		p1, _ := entity.NewPermission("dup.code", "Dup", nil, "res", entity.ActionRead)
		repo.Create(ctx, p1)

		p2, _ := entity.NewPermission("dup.code", "Dup 2", nil, "res", entity.ActionRead)
		err := repo.Create(ctx, p2)
		assert.Error(t, err)
	})
}

func TestPermissionRepository_GetByCode(t *testing.T) {
	db := setupPermissionDB(t)
	defer db.Close()

	repo := sqlite.NewPermissionRepository(db)
	ctx := context.Background()

	perm, _ := entity.NewPermission("test.get", "Test Get", nil, "test", entity.ActionRead)
	repo.Create(ctx, perm)

	found, err := repo.GetByCode(ctx, "test.get")
	require.NoError(t, err)
	assert.Equal(t, perm.ID, found.ID)
}
