package sqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"osbb-accounting/domain/entity"
	"osbb-accounting/infrastructure/persistence/sqlite"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSessionDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	query := `
	CREATE TABLE sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		ip_address TEXT,
		user_agent TEXT
	);
	`
	_, err = db.Exec(query)
	require.NoError(t, err)

	return db
}

func TestSessionRepository_Create(t *testing.T) {
	db := setupSessionDB(t)
	defer db.Close()

	repo := sqlite.NewSessionRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		session, err := entity.NewSession(1, time.Hour, nil, nil)
		require.NoError(t, err)

		err = repo.Create(ctx, session)
		require.NoError(t, err)
		assert.NotZero(t, session.ID)
	})
}

func TestSessionRepository_GetByToken(t *testing.T) {
	db := setupSessionDB(t)
	defer db.Close()

	repo := sqlite.NewSessionRepository(db)
	ctx := context.Background()

	session, _ := entity.NewSession(1, time.Hour, nil, nil)
	repo.Create(ctx, session)

	found, err := repo.GetByToken(ctx, session.Token)
	require.NoError(t, err)
	assert.Equal(t, session.ID, found.ID)
}

func TestSessionRepository_DeleteExpired(t *testing.T) {
	db := setupSessionDB(t)
	defer db.Close()

	repo := sqlite.NewSessionRepository(db)
	ctx := context.Background()

	// Expired session
	s1, _ := entity.NewSession(1, -time.Hour, nil, nil)
	// Manually set expiration to past because NewSession enforces min duration
	s1.ExpiresAt = time.Now().Add(-time.Hour)
	repo.Create(ctx, s1)

	// Valid session
	s2, _ := entity.NewSession(1, time.Hour, nil, nil)
	repo.Create(ctx, s2)

	count, err := repo.DeleteExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	_, err = repo.GetByID(ctx, s1.ID)
	assert.Error(t, err)

	_, err = repo.GetByID(ctx, s2.ID)
	require.NoError(t, err)
}
