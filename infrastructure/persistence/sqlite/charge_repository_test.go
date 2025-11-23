package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChargeRepository(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Initialize schema
	_, err = db.Exec(`
		CREATE TABLE charges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ownership_share_id INTEGER NOT NULL,
			charge_type TEXT NOT NULL,
			charge_date INTEGER NOT NULL,
			period_month INTEGER NOT NULL,
			period_year INTEGER NOT NULL,
			amount REAL NOT NULL,
			tariff REAL,
			quantity REAL,
			description TEXT,
			notes TEXT,
			deleted_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE ownership_shares (id INTEGER PRIMARY KEY, owner_id INTEGER, apartment_id INTEGER, share_numerator INTEGER, share_denominator INTEGER);
		CREATE TABLE owners (id INTEGER PRIMARY KEY, first_name TEXT, last_name TEXT, phone TEXT, email TEXT);
		CREATE TABLE apartments (id INTEGER PRIMARY KEY, apartment_number TEXT, floor INTEGER, entrance INTEGER);
	`)
	require.NoError(t, err)

	repo := NewChargeRepository(db)
	ctx := context.Background()

	// Setup test data
	_, err = db.Exec(`INSERT INTO owners (id, first_name, last_name) VALUES (1, 'John', 'Doe')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO apartments (id, apartment_number) VALUES (1, '101')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO ownership_shares (id, owner_id, apartment_id, share_numerator, share_denominator) VALUES (1, 1, 1, 1, 1)`)
	require.NoError(t, err)

	t.Run("Create and GetByID", func(t *testing.T) {
		charge := &entity.Charge{
			OwnershipShareID: 1,
			ChargeType:       entity.ChargeTypeMaintenance,
			ChargeDate:       time.Now(),
			PeriodMonth:      1,
			PeriodYear:       2024,
			Amount:           100.50,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		err := repo.Create(ctx, charge)
		assert.NoError(t, err)
		assert.NotZero(t, charge.ID)

		fetched, err := repo.GetByID(ctx, charge.ID)
		assert.NoError(t, err)
		assert.Equal(t, charge.ID, fetched.ID)
		assert.Equal(t, charge.Amount, fetched.Amount)
	})

	t.Run("Update", func(t *testing.T) {
		charge := &entity.Charge{
			OwnershipShareID: 1,
			ChargeType:       entity.ChargeTypeUtility,
			ChargeDate:       time.Now(),
			PeriodMonth:      2,
			PeriodYear:       2024,
			Amount:           200.00,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		err := repo.Create(ctx, charge)
		require.NoError(t, err)

		charge.Amount = 250.00
		err = repo.Update(ctx, charge)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, charge.ID)
		assert.NoError(t, err)
		assert.Equal(t, 250.00, fetched.Amount)
	})

	t.Run("SoftDelete and Restore", func(t *testing.T) {
		charge := &entity.Charge{
			OwnershipShareID: 1,
			ChargeType:       entity.ChargeTypeRepair,
			ChargeDate:       time.Now(),
			PeriodMonth:      3,
			PeriodYear:       2024,
			Amount:           300.00,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		err := repo.Create(ctx, charge)
		require.NoError(t, err)

		err = repo.SoftDelete(ctx, charge.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(ctx, charge.ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)

		err = repo.Restore(ctx, charge.ID)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, charge.ID)
		assert.NoError(t, err)
		assert.Equal(t, charge.ID, fetched.ID)
	})

	t.Run("List and Count", func(t *testing.T) {
		filter := repository.ChargeFilter{
			PeriodYear: func(i int) *int { return &i }(2024),
		}
		charges, err := repo.List(ctx, filter)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(charges), 3)

		count, err := repo.Count(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(len(charges)), count)
	})

	t.Run("CalculateTotalForPeriod", func(t *testing.T) {
		total, err := repo.CalculateTotalForPeriod(ctx, 1, 2024)
		assert.NoError(t, err)
		assert.Equal(t, 100.50, total)
	})
}
