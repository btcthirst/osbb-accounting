package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
	"osbb-accounting/infrastructure/persistence/sqlite"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupApartmentDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	query := `
	CREATE TABLE apartments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		apartment_number TEXT NOT NULL UNIQUE,
		floor INTEGER NOT NULL,
		entrance INTEGER,
		area_total REAL NOT NULL,
		area_living REAL,
		rooms_count INTEGER,
		cadastral_number TEXT,
		notes TEXT,
		is_active INTEGER NOT NULL DEFAULT 1,
		deleted_at INTEGER,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE TABLE ownership_shares (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		apartment_id INTEGER,
		is_active INTEGER,
		deleted_at INTEGER
	);
	`
	_, err = db.Exec(query)
	require.NoError(t, err)

	return db
}

func TestApartmentRepository_Create(t *testing.T) {
	db := setupApartmentDB(t)
	defer db.Close()

	repo := sqlite.NewApartmentRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		apt, err := entity.NewApartment("101", 1, 50.0, nil, nil, nil)
		require.NoError(t, err)

		err = repo.Create(ctx, apt)
		require.NoError(t, err)
		assert.NotZero(t, apt.ID)
	})

	t.Run("Duplicate Number", func(t *testing.T) {
		apt1, _ := entity.NewApartment("102", 1, 50.0, nil, nil, nil)
		repo.Create(ctx, apt1)

		apt2, _ := entity.NewApartment("102", 2, 60.0, nil, nil, nil)
		err := repo.Create(ctx, apt2)
		assert.Error(t, err)
	})
}

func TestApartmentRepository_GetByID(t *testing.T) {
	db := setupApartmentDB(t)
	defer db.Close()

	repo := sqlite.NewApartmentRepository(db)
	ctx := context.Background()

	apt, _ := entity.NewApartment("201", 2, 55.0, nil, nil, nil)
	repo.Create(ctx, apt)

	t.Run("Found", func(t *testing.T) {
		found, err := repo.GetByID(ctx, apt.ID)
		require.NoError(t, err)
		assert.Equal(t, apt.ApartmentNumber, found.ApartmentNumber)
	})

	t.Run("Not Found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 999)
		assert.Error(t, err)
	})
}

func TestApartmentRepository_List(t *testing.T) {
	db := setupApartmentDB(t)
	defer db.Close()

	repo := sqlite.NewApartmentRepository(db)
	ctx := context.Background()

	apt1, _ := entity.NewApartment("301", 3, 50.0, nil, nil, nil)
	repo.Create(ctx, apt1)
	apt2, _ := entity.NewApartment("302", 3, 60.0, nil, nil, nil)
	repo.Create(ctx, apt2)

	t.Run("List All", func(t *testing.T) {
		filter := repository.ApartmentFilter{}
		list, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("Filter by Number", func(t *testing.T) {
		filter := repository.ApartmentFilter{SearchQuery: "301"}
		list, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "301", list[0].ApartmentNumber)
	})
}
