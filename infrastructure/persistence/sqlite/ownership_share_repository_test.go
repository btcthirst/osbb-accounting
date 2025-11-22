package sqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
	"osbb-accounting/infrastructure/persistence/sqlite"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOwnershipShareDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	query := `
	CREATE TABLE ownership_shares (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		owner_id INTEGER NOT NULL,
		apartment_id INTEGER NOT NULL,
		share_numerator INTEGER NOT NULL,
		share_denominator INTEGER NOT NULL,
		ownership_type TEXT NOT NULL,
		start_date INTEGER NOT NULL,
		end_date INTEGER,
		document_type TEXT,
		document_number TEXT,
		document_date INTEGER,
		notes TEXT,
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

func TestOwnershipShareRepository_Create(t *testing.T) {
	db := setupOwnershipShareDB(t)
	defer db.Close()

	repo := sqlite.NewOwnershipShareRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		share, err := entity.NewOwnershipShare(1, 1, 50, 100, entity.OwnershipTypeFull, time.Now())
		require.NoError(t, err)

		err = repo.Create(ctx, share)
		require.NoError(t, err)
		assert.NotZero(t, share.ID)
	})

	t.Run("Exceeds 100%", func(t *testing.T) {
		// First share 60%
		s1, _ := entity.NewOwnershipShare(1, 2, 60, 100, entity.OwnershipTypeShared, time.Now())
		repo.Create(ctx, s1)

		// Second share 50% (Total 110%)
		s2, _ := entity.NewOwnershipShare(2, 2, 50, 100, entity.OwnershipTypeShared, time.Now())
		err := repo.Create(ctx, s2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceed 100%")
	})
}

func TestOwnershipShareRepository_List(t *testing.T) {
	db := setupOwnershipShareDB(t)
	defer db.Close()

	repo := sqlite.NewOwnershipShareRepository(db)
	ctx := context.Background()

	s1, _ := entity.NewOwnershipShare(1, 1, 50, 100, entity.OwnershipTypeShared, time.Now())
	repo.Create(ctx, s1)
	s2, _ := entity.NewOwnershipShare(2, 1, 50, 100, entity.OwnershipTypeShared, time.Now())
	repo.Create(ctx, s2)

	t.Run("List by Apartment", func(t *testing.T) {
		aptID := int64(1)
		filter := repository.OwnershipShareFilter{ApartmentID: &aptID}
		list, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})
}
