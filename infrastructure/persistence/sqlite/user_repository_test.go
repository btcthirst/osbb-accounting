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

func setupUserDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	query := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		middle_name TEXT,
		phone TEXT,
		is_active INTEGER NOT NULL DEFAULT 1,
		deleted_at INTEGER,
		last_login_at INTEGER,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE TABLE user_roles (
		user_id INTEGER NOT NULL,
		role_id INTEGER NOT NULL,
		PRIMARY KEY (user_id, role_id)
	);
	`
	_, err = db.Exec(query)
	require.NoError(t, err)

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupUserDB(t)
	defer db.Close()

	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		user, err := entity.NewUser("testuser", "test@example.com", "John", "Doe", nil, nil)
		require.NoError(t, err)
		user.PasswordHash = "password123"

		err = repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("Duplicate Username", func(t *testing.T) {
		user1, _ := entity.NewUser("dupuser", "email1@example.com", "John", "Doe", nil, nil)
		user1.PasswordHash = "pass"
		repo.Create(ctx, user1)

		user2, _ := entity.NewUser("dupuser", "email2@example.com", "Jane", "Doe", nil, nil)
		user2.PasswordHash = "pass"
		err := repo.Create(ctx, user2)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupUserDB(t)
	defer db.Close()

	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	user, _ := entity.NewUser("getid", "getid@example.com", "John", "Doe", nil, nil)
	user.PasswordHash = "pass"
	repo.Create(ctx, user)

	t.Run("Found", func(t *testing.T) {
		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Username, found.Username)
	})

	t.Run("Not Found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 999)
		assert.Error(t, err)
	})
}

func TestUserRepository_Update(t *testing.T) {
	db := setupUserDB(t)
	defer db.Close()

	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	user, _ := entity.NewUser("update", "update@example.com", "John", "Doe", nil, nil)
	user.PasswordHash = "pass"
	repo.Create(ctx, user)

	t.Run("Success", func(t *testing.T) {
		user.FirstName = "UpdatedName"
		err := repo.Update(ctx, user)
		require.NoError(t, err)

		updated, _ := repo.GetByID(ctx, user.ID)
		assert.Equal(t, "UpdatedName", updated.FirstName)
	})
}

func TestUserRepository_SoftDelete(t *testing.T) {
	db := setupUserDB(t)
	defer db.Close()

	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	user, _ := entity.NewUser("delete", "delete@example.com", "John", "Doe", nil, nil)
	user.PasswordHash = "pass"
	repo.Create(ctx, user)

	t.Run("Success", func(t *testing.T) {
		err := repo.SoftDelete(ctx, user.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, user.ID)
		assert.Error(t, err)
	})
}
