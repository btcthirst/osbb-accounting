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

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create owners table
	query := `
	CREATE TABLE owners (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		middle_name TEXT,
		phone TEXT,
		email TEXT,
		tax_number TEXT UNIQUE,
		passport_series TEXT,
		passport_number TEXT,
		registered_address TEXT,
		actual_address TEXT,
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

func TestOwnerRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := sqlite.NewOwnerRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		owner, err := entity.NewOwner("John", "Doe", nil, nil, nil, nil)
		require.NoError(t, err)

		err = repo.Create(ctx, owner)
		require.NoError(t, err)
		assert.NotZero(t, owner.ID)
		assert.WithinDuration(t, time.Now(), owner.CreatedAt, time.Second)
	})

	t.Run("Duplicate Tax Number", func(t *testing.T) {
		taxNumber := "1234567890"
		owner1, _ := entity.NewOwner("Jane", "Doe", nil, nil, nil, &taxNumber)
		err := repo.Create(ctx, owner1)
		require.NoError(t, err)

		owner2, _ := entity.NewOwner("John", "Smith", nil, nil, nil, &taxNumber)
		err = repo.Create(ctx, owner2)
		assert.Error(t, err)
		// Check if error is related to duplicate entry if possible,
		// though implementation might return wrapped error.
	})
}

func TestOwnerRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := sqlite.NewOwnerRepository(db)
	ctx := context.Background()

	owner, _ := entity.NewOwner("Alice", "Wonderland", nil, nil, nil, nil)
	err := repo.Create(ctx, owner)
	require.NoError(t, err)

	t.Run("Found", func(t *testing.T) {
		found, err := repo.GetByID(ctx, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, owner.ID, found.ID)
		assert.Equal(t, owner.FirstName, found.FirstName)
	})

	t.Run("Not Found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 999)
		assert.Error(t, err)
	})
}

func TestOwnerRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := sqlite.NewOwnerRepository(db)
	ctx := context.Background()

	owner, _ := entity.NewOwner("Bob", "Builder", nil, nil, nil, nil)
	err := repo.Create(ctx, owner)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		newName := "Robert"
		owner.FirstName = newName
		err := repo.Update(ctx, owner)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, newName, updated.FirstName)
	})
}

func TestOwnerRepository_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := sqlite.NewOwnerRepository(db)
	ctx := context.Background()

	owner, _ := entity.NewOwner("Charlie", "Chaplin", nil, nil, nil, nil)
	err := repo.Create(ctx, owner)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := repo.SoftDelete(ctx, owner.ID)
		require.NoError(t, err)

		// Should not be found by GetByID
		_, err = repo.GetByID(ctx, owner.ID)
		assert.Error(t, err)
	})
}

func TestOwnerRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := sqlite.NewOwnerRepository(db)
	ctx := context.Background()

	// Create test data
	names := []string{"Alice", "Bob", "Charlie"}
	for _, name := range names {
		owner, _ := entity.NewOwner(name, "Test", nil, nil, nil, nil)
		repo.Create(ctx, owner)
	}

	t.Run("List All", func(t *testing.T) {
		filter := repository.OwnerFilter{}
		owners, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, owners, 3)
	})

	t.Run("Search", func(t *testing.T) {
		filter := repository.OwnerFilter{SearchQuery: "Ali"}
		owners, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, owners, 1)
		assert.Equal(t, "Alice", owners[0].FirstName)
	})
}
