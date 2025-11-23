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

func TestPaymentRepository(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Initialize schema
	_, err = db.Exec(`
		CREATE TABLE payments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ownership_share_id INTEGER NOT NULL,
			payment_date INTEGER NOT NULL,
			amount REAL NOT NULL,
			payment_method TEXT NOT NULL,
			payment_purpose TEXT,
			period_month INTEGER,
			period_year INTEGER,
			receipt_number TEXT,
			notes TEXT,
			approved_by INTEGER,
			approved_at INTEGER,
			deleted_at INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE TABLE ownership_shares (id INTEGER PRIMARY KEY, owner_id INTEGER, apartment_id INTEGER, share_numerator INTEGER, share_denominator INTEGER);
		CREATE TABLE owners (id INTEGER PRIMARY KEY, first_name TEXT, last_name TEXT, phone TEXT, email TEXT);
		CREATE TABLE apartments (id INTEGER PRIMARY KEY, apartment_number TEXT, floor INTEGER, entrance INTEGER);
	`)
	require.NoError(t, err)

	repo := NewPaymentRepository(db)
	ctx := context.Background()

	// Setup test data
	_, err = db.Exec(`INSERT INTO owners (id, first_name, last_name) VALUES (1, 'John', 'Doe')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO apartments (id, apartment_number) VALUES (1, '101')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO ownership_shares (id, owner_id, apartment_id, share_numerator, share_denominator) VALUES (1, 1, 1, 1, 1)`)
	require.NoError(t, err)

	t.Run("Create and GetByID", func(t *testing.T) {
		payment := &entity.Payment{
			OwnershipShareID: 1,
			Amount:           500.00,
			PaymentMethod:    entity.PaymentMethodCard,
			PaymentPurpose:   "Utility payment",
			PaymentDate:      time.Now(),
			PeriodMonth:      func(i int) *int { return &i }(1),
			PeriodYear:       func(i int) *int { return &i }(2024),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		err := repo.Create(ctx, payment)
		assert.NoError(t, err)
		assert.NotZero(t, payment.ID)

		fetched, err := repo.GetByID(ctx, payment.ID)
		assert.NoError(t, err)
		assert.Equal(t, payment.ID, fetched.ID)
		assert.Equal(t, payment.Amount, fetched.Amount)
		assert.Equal(t, *payment.PeriodMonth, *fetched.PeriodMonth)
	})

	t.Run("Update", func(t *testing.T) {
		payment := &entity.Payment{
			OwnershipShareID: 1,
			Amount:           300.00,
			PaymentMethod:    entity.PaymentMethodCash,
			PaymentPurpose:   "Repair fund",
			PaymentDate:      time.Now(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		err := repo.Create(ctx, payment)
		require.NoError(t, err)

		payment.Amount = 350.00
		err = repo.Update(ctx, payment)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, payment.ID)
		assert.NoError(t, err)
		assert.Equal(t, 350.00, fetched.Amount)
	})

	t.Run("SoftDelete and Restore", func(t *testing.T) {
		payment := &entity.Payment{
			OwnershipShareID: 1,
			Amount:           100.00,
			PaymentMethod:    entity.PaymentMethodOther,
			PaymentPurpose:   "Donation",
			PaymentDate:      time.Now(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		err := repo.Create(ctx, payment)
		require.NoError(t, err)

		err = repo.SoftDelete(ctx, payment.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(ctx, payment.ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)

		err = repo.Restore(ctx, payment.ID)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, payment.ID)
		assert.NoError(t, err)
		assert.Equal(t, payment.ID, fetched.ID)
	})

	t.Run("List and Count", func(t *testing.T) {
		filter := repository.PaymentFilter{
			PeriodYear: func(i int) *int { return &i }(2024),
		}
		payments, err := repo.List(ctx, filter)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(payments), 1)

		count, err := repo.Count(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(len(payments)), count)
	})

	t.Run("GetTotalByPeriod", func(t *testing.T) {
		total, err := repo.GetTotalByPeriod(ctx, 1, 2024)
		assert.NoError(t, err)
		assert.Equal(t, 500.00, total)
	})
}
