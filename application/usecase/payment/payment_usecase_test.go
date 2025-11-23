package payment

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

// MockPermissionRepository is a simple mock for PermissionRepository
type MockPermissionRepository struct {
	repository.PermissionRepository
}

func (m *MockPermissionRepository) HasPermissionForResource(ctx context.Context, userID int64, resource string, action entity.Action) (bool, error) {
	return true, nil // Always allow in tests
}

func TestPaymentUseCases(t *testing.T) {
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
	`)
	require.NoError(t, err)

	paymentRepo := sqlite.NewPaymentRepository(db)
	permissionRepo := &MockPermissionRepository{}

	ctx := context.Background()

	t.Run("CreatePaymentUseCase", func(t *testing.T) {
		uc := NewCreatePaymentUseCase(paymentRepo, permissionRepo)
		input := CreatePaymentInput{
			CurrentUserID:    1,
			OwnershipShareID: 1,
			Amount:           500.00,
			PaymentMethod:    entity.PaymentMethodCard,
			PaymentPurpose:   "Utility payment",
			PaymentDate:      time.Now(),
		}

		output, err := uc.Execute(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, input.Amount, output.Amount)
	})

	t.Run("GetPaymentUseCase", func(t *testing.T) {
		// Create a payment first
		payment := &entity.Payment{
			OwnershipShareID: 1,
			Amount:           300.00,
			PaymentMethod:    entity.PaymentMethodCash,
			PaymentPurpose:   "Repair fund",
			PaymentDate:      time.Now(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		err := paymentRepo.Create(ctx, payment)
		require.NoError(t, err)

		uc := NewGetPaymentUseCase(paymentRepo, permissionRepo)
		input := GetPaymentInput{
			CurrentUserID: 1,
			PaymentID:     payment.ID,
		}

		output, err := uc.Execute(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, payment.ID, output.ID)
	})

	t.Run("ApprovePaymentUseCase", func(t *testing.T) {
		// Create a payment first
		payment := &entity.Payment{
			OwnershipShareID: 1,
			Amount:           100.00,
			PaymentMethod:    entity.PaymentMethodOther,
			PaymentPurpose:   "Donation",
			PaymentDate:      time.Now(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		err := paymentRepo.Create(ctx, payment)
		require.NoError(t, err)

		uc := NewApprovePaymentUseCase(paymentRepo, permissionRepo)
		input := ApprovePaymentInput{
			CurrentUserID: 2, // Approver ID
			PaymentID:     payment.ID,
		}

		output, err := uc.Execute(ctx, input)
		assert.NoError(t, err)
		assert.True(t, output.IsApproved)
		assert.Equal(t, int64(2), *output.ApprovedBy)
	})
}
