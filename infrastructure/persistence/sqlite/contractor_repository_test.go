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

func setupContractorDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	query := `
	CREATE TABLE contractors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		edrpou TEXT UNIQUE,
		contractor_type TEXT NOT NULL,
		contact_person TEXT,
		phone TEXT,
		email TEXT,
		address TEXT,
		bank_account TEXT,
		bank_name TEXT,
		bank_mfo TEXT,
		contract_number TEXT,
		contract_date INTEGER,
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

func TestContractorRepository_Create(t *testing.T) {
	db := setupContractorDB(t)
	defer db.Close()

	repo := sqlite.NewContractorRepository(db)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		contractor, err := entity.NewContractor("Test Contractor", entity.ContractorTypeSupplier, nil, nil, nil, nil)
		require.NoError(t, err)

		err = repo.Create(ctx, contractor)
		require.NoError(t, err)
		assert.NotZero(t, contractor.ID)
	})

	t.Run("Duplicate EDRPOU", func(t *testing.T) {
		edrpou := "12345678"
		c1, _ := entity.NewContractor("C1", entity.ContractorTypeSupplier, &edrpou, nil, nil, nil)
		repo.Create(ctx, c1)

		c2, _ := entity.NewContractor("C2", entity.ContractorTypeSupplier, &edrpou, nil, nil, nil)
		err := repo.Create(ctx, c2)
		assert.Error(t, err)
	})
}

func TestContractorRepository_List(t *testing.T) {
	db := setupContractorDB(t)
	defer db.Close()

	repo := sqlite.NewContractorRepository(db)
	ctx := context.Background()

	c1, _ := entity.NewContractor("Alpha", entity.ContractorTypeSupplier, nil, nil, nil, nil)
	repo.Create(ctx, c1)
	c2, _ := entity.NewContractor("Beta", entity.ContractorTypeService, nil, nil, nil, nil)
	repo.Create(ctx, c2)

	t.Run("List All", func(t *testing.T) {
		filter := repository.ContractorFilter{}
		list, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("Filter by Type", func(t *testing.T) {
		typ := entity.ContractorTypeService
		filter := repository.ContractorFilter{ContractorType: &typ}
		list, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "Beta", list[0].Name)
	})
}
