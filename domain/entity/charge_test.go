package entity_test

import (
	"testing"
	"time"

	"osbb-accounting/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCharge(t *testing.T) {
	now := time.Now()
	tariff := 10.0
	quantity := 5.0

	t.Run("Success", func(t *testing.T) {
		charge, err := entity.NewCharge(
			1,
			entity.ChargeTypeMaintenance,
			now,
			int(now.Month()),
			now.Year(),
			50.0,
			&tariff,
			&quantity,
			nil,
		)
		require.NoError(t, err)
		assert.NotNil(t, charge)
		assert.Equal(t, 50.0, charge.Amount)
	})

	t.Run("Validation Errors", func(t *testing.T) {
		tests := []struct {
			name    string
			amount  float64
			tariff  *float64
			qty     *float64
			wantErr error
		}{
			{"Invalid Amount", -10, nil, nil, entity.ErrChargeAmountInvalid},
			{"Inconsistent Calculation", 40.0, &tariff, &quantity, entity.ErrChargeCalculationInconsistent},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := entity.NewCharge(
					1,
					entity.ChargeTypeMaintenance,
					now,
					int(now.Month()),
					now.Year(),
					tt.amount,
					tt.tariff,
					tt.qty,
					nil,
				)
				assert.ErrorIs(t, err, tt.wantErr)
			})
		}
	})
}

func TestCharge_RecalculateAmount(t *testing.T) {
	tariff := 10.0
	quantity := 5.0
	charge, _ := entity.NewCharge(1, entity.ChargeTypeMaintenance, time.Now(), 1, 2024, 50.0, &tariff, &quantity, nil)

	t.Run("Success", func(t *testing.T) {
		newTariff := 12.0
		charge.Tariff = &newTariff
		err := charge.RecalculateAmount()
		require.NoError(t, err)
		assert.Equal(t, 60.0, charge.Amount)
	})

	t.Run("Missing Data", func(t *testing.T) {
		charge.Tariff = nil
		err := charge.RecalculateAmount()
		assert.Error(t, err)
	})
}

func TestCharge_CalculatePenalty(t *testing.T) {
	now := time.Now()
	// Create a charge for a period 2 months ago
	pastDate := now.AddDate(0, -2, 0)
	charge, _ := entity.NewCharge(
		1,
		entity.ChargeTypeMaintenance,
		pastDate,
		int(pastDate.Month()),
		pastDate.Year(),
		1000.0,
		nil,
		nil,
		nil,
	)

	t.Run("Overdue Penalty", func(t *testing.T) {
		// 0.1% per day
		penalty := charge.CalculatePenalty(0.001)
		assert.Greater(t, penalty, 0.0)
	})

	t.Run("No Penalty for Penalty Type", func(t *testing.T) {
		charge.ChargeType = entity.ChargeTypePenalty
		penalty := charge.CalculatePenalty(0.001)
		assert.Equal(t, 0.0, penalty)
	})
}
