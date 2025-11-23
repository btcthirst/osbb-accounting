package charge

import (
	"context"
	"testing"
	"time"

	"osbb-accounting/domain/entity"
	"osbb-accounting/domain/repository"
	"osbb-accounting/infrastructure/persistence/sqlite"

	"database/sql"

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

func TestChargeUseCases(t *testing.T) {
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
	`)
	require.NoError(t, err)

	chargeRepo := sqlite.NewChargeRepository(db)
	permissionRepo := &MockPermissionRepository{}

	ctx := context.Background()

	t.Run("CreateChargeUseCase", func(t *testing.T) {
		uc := NewCreateChargeUseCase(chargeRepo, permissionRepo)
		input := CreateChargeInput{
			CurrentUserID:    1,
			OwnershipShareID: 1,
			ChargeType:       entity.ChargeTypeMaintenance,
			ChargeDate:       time.Now(),
			PeriodMonth:      1,
			PeriodYear:       2024,
			Amount:           100.00,
		}

		output, err := uc.Execute(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, input.Amount, output.Amount)
	})

	t.Run("GetChargeUseCase", func(t *testing.T) {
		// Create a charge first
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
		err := chargeRepo.Create(ctx, charge)
		require.NoError(t, err)

		uc := NewGetChargeUseCase(chargeRepo, permissionRepo)
		input := GetChargeInput{
			CurrentUserID: 1,
			ChargeID:      charge.ID,
		}

		output, err := uc.Execute(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, charge.ID, output.ID)
	})

	t.Run("ListChargesUseCase", func(t *testing.T) {
		uc := NewListChargesUseCase(chargeRepo, permissionRepo)
		input := ListChargesInput{
			CurrentUserID: 1,
			Limit:         10,
		}

		output, err := uc.Execute(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.GreaterOrEqual(t, len(output.Charges), 2)
	})
}
